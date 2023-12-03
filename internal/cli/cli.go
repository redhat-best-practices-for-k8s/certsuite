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

const banner = `

 _____  _   _ ______  _____  _____ ______  _____    __        _____     _____ __  
/  __ \| \ | ||  ___|/  __ \|  ___|| ___ \|_   _|  / /       |  ___|   |  _  |\ \ 
| /  \/|  \| || |_   | /  \/| |__  | |_/ /  | |   | | __   __|___ \    | |/' | | |
| |    | .   ||  _|  | |    |  __| |    /   | |   | | \ \ / /    \ \   |  /| | | |
| \__/\| |\  || |    | \__/\| |___ | |\ \   | |   | |  \ V / /\__/ / _ \ |_/ / | |
 \____/\_| \_/\_|     \____/\____/ \_| \_|  \_/   | |   \_/  \____/ (_) \___/  | |
                                                   \_\                        /_/ 
																				 

`

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
