package cmd

import (
	"fmt"
	"os"
	"stash/src/actions"
	"strconv"

	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
)


func init() {
  // TODO
  // - flatten tree
  // - individual apply
  // - restore

	rootCmd.AddCommand(apply)
  rootCmd.AddCommand(version)
  rootCmd.AddCommand(restore)
  rootCmd.AddCommand(revert)

  initArgs()
}



var version = &cobra.Command{
	Use:   "version",
	Short: "Prints app version",
	Long:  "Prints app version",
	Run: func(cmd *cobra.Command, args []string) {
		actions.PrintVersion()
	},
}

var apply = &cobra.Command{
	Use:   "apply",
	Short: "Applies configs to the system",
	Long:  "Applies configs to the system",
	Run: func(cmd *cobra.Command, args []string) {
		actions.Apply()
	},
}

var revert = &cobra.Command{
	Use:   "revert",
	Short: "reverts back to original state",
	Long:  "reverts back to original state",
	Run: func(cmd *cobra.Command, args []string) {
      actions.Revert()
	},
}

var restore_list bool

var restore = &cobra.Command{
	Use:   "restore",
	Short: "restores to a previous state",
	Long:  "restores to a previous state",
	Run: func(cmd *cobra.Command, args []string) {
    if restore_list {
      actions.PrintRestorePoints()
    } else {
      if len(args) == 0 {
        fmt.Println("No restore index param provided!")
        os.Exit(1)
      }

      // index, _ := strconv.ParseInt(args[0], 10, 0)
      index, _ := strconv.Atoi(args[0])
      actions.Restore(index)
    }
	},
}

func initArgs() {
  // restore args
  // var list string
  restore.Flags().BoolVarP(&restore_list, "list", "l", false, "Displays list of restore points")

}


var rootCmd = &cobra.Command{
	Use:   "stash",
	Short: "Declaration system configuration manager",
	Long: "Declaration system configuration manager",
	Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Created by: Dom Di Nenno")
    actions.PrintVersion()
    fmt.Println("See --help for more info")
	},
}
  
func Execute() {


	// fonts := []string{
    //     "3-d",
    //     "3x5",
    //     "5lineoblique",
    //     "acrobatic",
    //     "alligator",
    //     "alligator2",
    //     "alphabet",
    //     "avatar",
    //     "banner",
    //     "banner3-D",
    //     "banner3",
    //     "banner4",
    //     "barbwire",
    //     "basic",
    //     "bell",
    //     "big",
    //     "bigchief",
    //     "binary",
    //     "block",
    //     "bubble",
    //     "bulbhead",
    //     "calgphy2",
    //     "caligraphy",
    //     "catwalk",
    //     "chunky",
    //     "coinstak",
    //     "colossal",
    //     "computer",
    //     "contessa",
    //     "contrast",
    //     "cosmic",
    //     "cosmike",
    //     "cricket",
    //     "cursive",
    //     "cyberlarge",
    //     "cybermedium",
    //     "cybersmall",
    //     "diamond",
    //     "digital",
    //     "doh",
    //     "doom",
    //     "dotmatrix",
    //     "drpepper",
    //     "eftichess",
    //     "eftifont",
    //     "eftipiti",
    //     "eftirobot",
    //     "eftitalic",
    //     "eftiwall",
    //     "eftiwater",
    //     "epic",
    //     "fender",
    //     "fourtops",
    //     "fuzzy",
    //     "goofy",
    //     "gothic",
    //     "graffiti",
    //     "hollywood",
    //     "invita",
    //     "isometric1",
    //     "isometric2",
    //     "isometric3",
    //     "isometric4",
    //     "italic",
    //     "ivrit",
    //     "jazmine",
    //     "jerusalem",
    //     "katakana",
    //     "kban",
    //     "larry3d",
    //     "lcd",
    //     "lean",
    //     "letters",
    //     "linux",
    //     "lockergnome",
    //     "madrid",
    //     "marquee",
    //     "maxfour",
    //     "mike",
    //     "mini",
    //     "mirror",
    //     "mnemonic",
    //     "morse",
    //     "moscow",
    //     "nancyj-fancy",
    //     "nancyj-underlined",
    //     "nancyj",
    //     "nipples",
    //     "ntgreek",
    //     "o8",
    //     "ogre",
    //     "pawp",
    //     "peaks",
    //     "pebbles",
    //     "pepper",
    //     "poison",
    //     "puffy",
    //     "pyramid",
    //     "rectangles",
    //     "relief",
    //     "relief2",
    //     "rev",
    //     "roman",
    //     "rot13",
    //     "rounded",
    //     "rowancap",
    //     "rozzo",
    //     "runic",
    //     "runyc",
    //     "sblood",
    //     "script",
    //     "serifcap",
    //     "shadow",
    //     "short",
    //     "slant",
    //     "slide",
    //     "slscript",
    //     "small",
    //     "smisome1",
    //     "smkeyboard",
    //     "smscript",
    //     "smshadow",
    //     "smslant",
    //     "smtengwar",
    //     "speed",
    //     "stampatello",
    //     "standard",
    //     "starwars",
    //     "stellar",
    //     "stop",
    //     "straight",
    //     "tanja",
    //     "tengwar",
    //     "term",
    //     "thick",
    //     "thin",
    //     "threepoint",
    //     "ticks",
    //     "ticksslant",
    //     "tinker-toy",
    //     "tombstone",
    //     "trek",
    //     "tsalagi",
    //     "twopoint",
    //     "univers",
    //     "usaflag",
    //     "wavy",
    //     "weird",
    // }


	// sblood
	// alligator2
	// alligator
	// computer
	// cricket
	// larry3d

	// for i := 0; i < len(fonts); i++ {
	// 	// font := fonts[i]
	// 	// test := figure.NewColorFigure("STASH", font, "cyan", true)
	// 	// test.Print()

	// 	// fmt.Println(font)
	// }

	// Print Logo
	figure.NewColorFigure("STASH", "cricket", "yellow", true).Print()
  fmt.Println("")
	
  if err := rootCmd.Execute(); err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }
}