package cli

import (
	"fmt"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

// ASCII art generated on http://www.patorjk.com/software/taag/ with
// the font DOOM by Frans P. de Vries <fpv@xymph.iaf.nl>  18 Jun 1996.
// All backticks (`) were removed for string literal compatibility.
const banner = `

 _____  _   _ ______  _____  _____ ______  _____    __        _____     _____ __  
/  __ \| \ | ||  ___|/  __ \|  ___|| ___ \|_   _|  / /       |  ___|   |  _  |\ \ 
| /  \/|  \| || |_   | /  \/| |__  | |_/ /  | |   | | __   __|___ \    | |/' | | |
| |    | .   ||  _|  | |    |  __| |    /   | |   | | \ \ / /    \ \   |  /| | | |
| \__/\| |\  || |    | \__/\| |___ | |\ \   | |   | |  \ V / /\__/ / _ \ |_/ / | |
 \____/\_| \_/\_|     \____/\____/ \_| \_|  \_/   | |   \_/  \____/ (_) \___/  | |
                                                   \_\                        /_/ 
																				 

`

const (
	CheckResultTagPass    = Green + "PASS" + Reset
	CheckResultTagFail    = Red + "FAIL" + Reset
	CheckResultTagSkip    = Yellow + "SKIP" + Reset
	CheckResultTagRunning = Cyan + "RUNNING" + Reset
	CheckResultTagAborted = Red + "ABORTED" + Reset
)

func PrintBanner() {
	fmt.Print(banner)
}

func PrintResultsTable(results map[string][]int) {
	fmt.Printf("\n")
	fmt.Println("-----------------------------------------------------------")
	fmt.Printf("| %-27s %-9s %-9s %s |\n", "SUITE", "PASSED", "FAILED", "SKIPPED")
	fmt.Println("-----------------------------------------------------------")
	for groupName, groupResults := range results {
		fmt.Printf("| %-25s %8d %9d %10d |\n", groupName,
			groupResults[0],
			groupResults[1],
			groupResults[2])
		fmt.Println("-----------------------------------------------------------")
	}
	fmt.Printf("\n")
}
