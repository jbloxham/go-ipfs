package commands

import (
	"fmt"

	cmds "gx/ipfs/Qma6uuSyjkecGhMFFLfzyJDPyoDtNJSHJNweDccZhaWkgU/go-ipfs-cmds"
	"gx/ipfs/Qmde5VP1qUkyQXKCfmEUA7bP64V2HAptbJ7phuPp7jXWwg/go-ipfs-cmdkit"
)

var IndexCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Index a file for searching on IPFS.",
		ShortDescription: `
'ipfs index' will index yo files.
		`,
	},
	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("ipfs-path", true, true, "The path to the IPFS object to be indexed.").EnableStdin(),
	},
	Options: []cmdkit.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		return fmt.Errorf("The command is working!!!!! :) :)")
	},
	//PostRun: cmds.PostRunMap{},
}