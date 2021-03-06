package main

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/republicprotocol/co-go"
	"github.com/urfave/cli"
)

// updateNode updates the Darknode to the latest release. It can also be used
// to update the config file of the darknode.
func updateNode(ctx *cli.Context) error {
	name := ctx.Args().First()
	updateConfig := ctx.Bool("config")
	tags := ctx.String("tags")
	branch := ctx.String("branch")

	if tags == "" && name == "" {
		return ErrEmptyNodeName
	} else if tags == "" && name != "" {
		return updateSingleNode(name, branch, updateConfig)
	} else if tags != "" && name == "" {
		nodes, err := getNodesByTags(tags)
		if err != nil {
			return err
		}
		errs := make([]error, len(nodes))
		co.ForAll(nodes, func(i int) {
			errs[i] = updateSingleNode(nodes[i], branch, updateConfig)
		})
		return handleErrs(errs)
	}

	return ErrNameAndTags
}

func updateSingleNode(name, branch string, updateConfig bool) error {
	nodePath := nodePath(name)
	keyPairPath := path.Join(nodePath, "ssh_keypair")
	ip, err := getIp(nodePath)
	if err != nil {
		return err
	}

	// Check if we need to update the node config
	if updateConfig {
		data, err := ioutil.ReadFile(path.Join(nodePath, "config.json"))
		if err != nil {
			return err
		}
		updateConfigScript := fmt.Sprintf(`echo '%s' > $HOME/.darknode/config.json`, string(data))
		if err := run("ssh", "-i", keyPairPath, "darknode@"+ip, "-oStrictHostKeyChecking=no", updateConfigScript); err != nil {
			return err
		}

		fmt.Printf("%sConfig of [%s] has been updated to the local version.%s\n", GREEN, name, RESET)
	}

	udpate, err := ioutil.ReadFile(path.Join(Directory, "scripts", "update.sh"))
	if err != nil {
		return err
	}
	err = run("ssh", "-i", keyPairPath, "darknode@"+ip, "-oStrictHostKeyChecking=no", string(udpate))
	if err != nil {
		return err
	}
	fmt.Printf("%s[%s] has been updated to the latest version on %s branch.%s \n", GREEN, name, branch, RESET)
	return nil
}

// sshNode will ssh into the Darknode
func sshNode(ctx *cli.Context) error {
	name := ctx.Args().First()
	if name == "" {
		cli.ShowCommandHelp(ctx, "ssh")
		return ErrEmptyNodeName
	}
	nodePath := nodePath(name)
	ip, err := getIp(nodePath)
	if err != nil {
		return err
	}
	keyPairPath := nodePath + "/ssh_keypair"

	return run("ssh", "-i", keyPairPath, "darknode@"+ip)
}
