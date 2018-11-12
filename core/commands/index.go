package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
	"github.com/ipfs/go-ipfs/core/coreapi/interface"

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
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}
		
		return index(req.Context, api, req.Arguments)
	},
}

func index(ctx context.Context, api iface.CoreAPI, paths []string) error {
	index := make(map[string]map[string]float64)
	for _, p := range paths {
		fpath, err := iface.ParsePath(p)
		if err != nil {
			return err
		}

		file, err := api.Unixfs().Get(ctx, fpath)
		if err != nil {
			return err
		}

		if file.IsDirectory() {
			return iface.ErrIsDir
		}

		var data []byte
		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			data = append(data, buf[:n]...)
			if err == io.EOF {
				break
			}
		}

		keywords := strings.Fields(string(data))

		for _, keyword := range keywords {
			_, prs := index[keyword]
			if prs {
				index[keyword][p] += 1
			} else {
				index[keyword] = map[string]float64{p:1}
			}
		}
	}

	jsonIndex, err := json.Marshal(index)
	if err != nil {
		return err
	}

	return fmt.Errorf(string(jsonIndex))
}