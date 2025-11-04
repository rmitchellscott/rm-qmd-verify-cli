package display

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/api"
)

var (
	titleStyle = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(1)

	compatibleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	incompatibleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)

	noDataStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	headerStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)

	cellStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)

	versionCellStyle = lipgloss.NewStyle().
				Align(lipgloss.Left)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Italic(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF"))
)

type matrixCell struct {
	compatible bool
	hasData    bool
	errorDetail string
}

func RenderComparisonResults(response *api.ComparisonResponse, verbose bool) {
	matrix := buildCompatibilityMatrix(response)
	devices := getDeviceOrder(matrix)
	versions := getSortedVersions(matrix)

	if len(versions) == 0 {
		fmt.Println(infoStyle.Render("No compatibility data available"))
		return
	}

	tableStr := buildMatrixTable(matrix, versions, devices, verbose)

	title := "QMD Compatibility Check Results"
	titleWidth := lipgloss.Width(tableStr)
	centeredTitle := lipgloss.NewStyle().
		Width(titleWidth).
		Align(lipgloss.Center).
		Render(title)

	fmt.Println()
	fmt.Println(centeredTitle)
	fmt.Println()
	fmt.Println(tableStr)
	fmt.Println()

	var compatibleCount string
	if len(response.Compatible) > 0 {
		compatibleCount = compatibleStyle.Render(fmt.Sprintf("%d compatible", len(response.Compatible)))
	} else {
		compatibleCount = fmt.Sprintf("%d compatible", len(response.Compatible))
	}

	var incompatibleCount string
	if len(response.Incompatible) > 0 {
		incompatibleCount = incompatibleStyle.Render(fmt.Sprintf("%d incompatible", len(response.Incompatible)))
	} else {
		incompatibleCount = fmt.Sprintf("%d incompatible", len(response.Incompatible))
	}

	summary := fmt.Sprintf("Summary: %d checked | %s | %s",
		response.TotalChecked,
		compatibleCount,
		incompatibleCount)
	fmt.Println(summary)
}

func buildMatrixTable(matrix map[string]map[string]matrixCell, versions []string, devices []string, verbose bool) string {
	var output strings.Builder

	deviceColWidth := 6
	for _, device := range devices {
		if len(device) > deviceColWidth {
			deviceColWidth = len(device)
		}
	}

	versionColWidth := 15
	for _, version := range versions {
		if len(version) > versionColWidth {
			versionColWidth = len(version)
		}
	}

	renderMatrixHeaderToBuilder(&output, devices, versionColWidth, deviceColWidth)
	renderMatrixSeparatorToBuilder(&output, len(devices), versionColWidth, deviceColWidth)

	var errorDetails []string

	for _, version := range versions {
		deviceRow := matrix[version]

		versionCell := versionCellStyle.Width(versionColWidth).Render(version)
		output.WriteString(" " + versionCell + " ")

		for _, device := range devices {
			cell, exists := deviceRow[device]
			var content string
			if !exists || !cell.hasData {
				content = noDataStyle.Render("—")
			} else if cell.compatible {
				content = compatibleStyle.Render("✓")
			} else {
				content = incompatibleStyle.Render("✗")
				if verbose && cell.errorDetail != "" {
					errorDetails = append(errorDetails, fmt.Sprintf("%s (%s): %s",
						version, device, cell.errorDetail))
				}
			}

			cellRendered := cellStyle.Width(deviceColWidth).Render(content)
			output.WriteString(cellRendered)
		}
		output.WriteString("\n")
	}

	if verbose && len(errorDetails) > 0 {
		output.WriteString("\n")
		output.WriteString(errorStyle.Render("Error Details:") + "\n")
		for _, detail := range errorDetails {
			output.WriteString(errorStyle.Render("  • "+detail) + "\n")
		}
	}

	return output.String()
}

func renderMatrixHeaderToBuilder(output *strings.Builder, devices []string, versionColWidth, deviceColWidth int) {
	versionHeader := versionCellStyle.Width(versionColWidth).Render("")
	output.WriteString(" " + versionHeader + " ")

	for _, device := range devices {
		headerCell := headerStyle.Width(deviceColWidth).Render(device)
		output.WriteString(headerCell)
	}
	output.WriteString("\n")
}

func renderMatrixSeparatorToBuilder(output *strings.Builder, deviceCount, versionColWidth, deviceColWidth int) {
	totalWidth := versionColWidth + 2 + (deviceColWidth * deviceCount)
	output.WriteString(strings.Repeat("─", totalWidth))
	output.WriteString("\n")
}

func buildCompatibilityMatrix(response *api.ComparisonResponse) map[string]map[string]matrixCell {
	matrix := make(map[string]map[string]matrixCell)

	for _, result := range response.Compatible {
		if matrix[result.OSVersion] == nil {
			matrix[result.OSVersion] = make(map[string]matrixCell)
		}
		matrix[result.OSVersion][result.Device] = matrixCell{
			compatible: true,
			hasData:    true,
		}
	}

	for _, result := range response.Incompatible {
		if matrix[result.OSVersion] == nil {
			matrix[result.OSVersion] = make(map[string]matrixCell)
		}
		matrix[result.OSVersion][result.Device] = matrixCell{
			compatible:  false,
			hasData:     true,
			errorDetail: result.ErrorDetail,
		}
	}

	return matrix
}

func getDeviceOrder(matrix map[string]map[string]matrixCell) []string {
	deviceSet := make(map[string]bool)
	for _, devices := range matrix {
		for device := range devices {
			deviceSet[device] = true
		}
	}

	devices := make([]string, 0, len(deviceSet))
	for device := range deviceSet {
		devices = append(devices, device)
	}

	deviceOrder := map[string]int{
		"rm1":   0,
		"rm2":   1,
		"rmpp":  2,
		"rmppm": 3,
	}

	sort.Slice(devices, func(i, j int) bool {
		orderI, okI := deviceOrder[devices[i]]
		orderJ, okJ := deviceOrder[devices[j]]
		if okI && okJ {
			return orderI < orderJ
		}
		if okI {
			return true
		}
		if okJ {
			return false
		}
		return devices[i] < devices[j]
	})

	return devices
}

func getSortedVersions(matrix map[string]map[string]matrixCell) []string {
	versions := make([]string, 0, len(matrix))
	for version := range matrix {
		versions = append(versions, version)
	}

	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) > 0
	})

	return versions
}

func compareVersions(v1, v2 string) int {
	p1 := strings.Split(v1, ".")
	p2 := strings.Split(v2, ".")

	maxLen := len(p1)
	if len(p2) > maxLen {
		maxLen = len(p2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(p1) {
			fmt.Sscanf(p1[i], "%d", &n1)
		}
		if i < len(p2) {
			fmt.Sscanf(p2[i], "%d", &n2)
		}

		if n1 != n2 {
			return n1 - n2
		}
	}

	return 0
}

func RenderHashtableList(response *api.HashtablesResponse) {
	fmt.Println(titleStyle.Render("Available Hashtables"))
	fmt.Println()

	if response.Count == 0 {
		fmt.Println(infoStyle.Render("No hashtables available on the server"))
		return
	}

	headers := []string{"Device", "OS Version", "Hashtable", "Entries"}
	colWidths := []int{10, 12, 25, 10}

	for _, ht := range response.Hashtables {
		if len(ht.Device) > colWidths[0] {
			colWidths[0] = len(ht.Device)
		}
		if len(ht.OSVersion) > colWidths[1] {
			colWidths[1] = len(ht.OSVersion)
		}
		if len(ht.Name) > colWidths[2] {
			colWidths[2] = len(ht.Name)
		}
	}

	renderTableHeader(headers, colWidths)
	renderTableSeparator(colWidths)

	for _, ht := range response.Hashtables {
		row := []string{
			ht.Device,
			ht.OSVersion,
			ht.Name,
			fmt.Sprintf("%d", ht.EntryCount),
		}
		renderTableRow(row, colWidths)
	}

	fmt.Println()
	fmt.Printf("Total Hashtables: %d\n", response.Count)
}

func renderTableHeader(headers []string, widths []int) {
	var cells []string
	for i, header := range headers {
		cell := lipgloss.NewStyle().Width(widths[i]).Render(header)
		cells = append(cells, cell)
	}
	fmt.Println(" " + lipgloss.JoinHorizontal(lipgloss.Left, cells...))
}

func renderTableSeparator(widths []int) {
	totalWidth := 0
	for _, width := range widths {
		totalWidth += width
	}
	fmt.Println(strings.Repeat("─", totalWidth+len(widths)))
}

func renderTableRow(cells []string, widths []int) {
	var renderedCells []string
	for i, cell := range cells {
		rendered := lipgloss.NewStyle().Width(widths[i]).Render(cell)
		renderedCells = append(renderedCells, rendered)
	}
	fmt.Println(" " + lipgloss.JoinHorizontal(lipgloss.Left, renderedCells...))
}

func RenderError(err error) {
	fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %s", err.Error())))
}

func RenderSuccess(message string) {
	fmt.Println(compatibleStyle.Render(message))
}

func RenderInfo(message string) {
	fmt.Println(infoStyle.Render(message))
}
