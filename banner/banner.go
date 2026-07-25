package banner

import (
	"fmt"
)

// prints the version message
const version = "v0.0.2"

func PrintVersion() {
	fmt.Printf("Current indextree version %s\n", version)
}

// Prints the Colorful banner
func PrintBanner() {
	banner := `                                        
    _             __             __                  
   (_)____   ____/ /___   _  __ / /_ _____ ___   ___ 
  / // __ \ / __  // _ \ | |/_// __// ___// _ \ / _ \
 / // / / // /_/ //  __/_>  < / /_ / /   /  __//  __/
/_//_/ /_/ \__,_/ \___//_/|_| \__//_/    \___/ \___/
`
	fmt.Printf("%s\n%40s\n\n", banner, "Current indextree version "+version)
}
