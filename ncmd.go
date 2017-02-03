package ncmd

import (
	"fmt"
	"log"
	"os/exec"
)

type NCmd struct {
	Out string
}

//run ssh command
func (u *NCmd) Execute(cmd string, args ...string) error {

	log.Println("log:", cmd, args)

	out, err := exec.Command(cmd, args...).Output()

	if err != nil {
		log.Println("Error running:", err.Error())
		return err
	}

	u.Out = fmt.Sprintf("%s", out)

	return nil
}
