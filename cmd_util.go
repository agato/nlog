package cmd

import (
	"fmt"
	"os/exec"
	"log"
)

type CmdUtil struct {
	Out  string
}



//run ssh command
func (u *CmdUtil) Execute(cmd string, args ...string) error {

	log.Println("log:", cmd, args)

	out,err := exec.Command(cmd, args...).Output()

	if err != nil {
		log.Println("Error running:", err.Error())
		return err
	}

	u.Out = fmt.Sprintf("%s", out)

	return nil
}
