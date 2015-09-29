package dsl

import (
	"fmt"
	"github.com/tubs-ips/nutsh/cli"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	cmdline                 cli.CLI
	lastCommand, lastOutput string
	wasInteractive          bool
	didOutput               bool
	running                 bool
	lastOutputWasSay        bool
)

func init() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		for {
			<-c
			if running {
				cmdline.Interrupt()
				select {
				case <-time.After(time.Second):
					break
				case <-c:
					cmdline.Interrupt()
					Say("Geben Sie zum Beenden der Nut Shell `exit` ein.")
				}
			}
		}
	}()
	cli.UseStdin()
}

func Spawn(target string) {
	cmdline = cli.Spawn(target)
	running = true
}

func Query(query string) (string, bool) {
	s, ok := cmdline.Query(" " + query)
	if ! ok {
		return "", false
	}
	return strings.TrimSpace(s), true
}

func SimulatePrompt(query string, interaction string) bool {
	lastCommand = query
	fmt.Println("$ " + query)
	var ok bool
	lastOutput, ok = cmdline.QueryInteractive(query, interaction)
	if ! ok {
		return false
	}
	fmt.Print(lastOutput)
	lastOutputWasSay = false
	return true
}

func QueryOutput(query string, expression string) (bool, bool) {
	output, ok := cmdline.Query(query)
	if ! ok {
		return false, false
	}
	return regexp.MustCompile(expression).MatchString(output), true
}

func Say(text string) {
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "[32m$1[36m")
	text = regexp.MustCompile("(^|[^\\\\])\\*([^*]+[^\\\\])\\*").ReplaceAllString(text, "$1[33m$2[36m")
	text = regexp.MustCompile("\\\\\\*").ReplaceAllString(text, "*")
	text = regexp.MustCompile("\\s+").ReplaceAllString(text, " ")
	_, c := getsize()
	if ! lastOutputWasSay {
		fmt.Printf("\n")
	}
	fmt.Printf("[36m%s\n\n[0m", indent(wrap(text, c-4), 4))
	lastOutputWasSay = true
}

func wrap(text string, width int) string {
	ret := ""
	line_len := 0
	for _, w := range strings.Split(text, " ") {
		l := len(w)
		if line_len + l + 1 > width {
			ret += "\n"
			line_len = 0
		}
		ret += w+" "
		line_len += l + 1
	}
	return ret
}

func indent(text string, spaces int) string {
	iden := ""
	for i := 0; i < spaces; i++ {
		iden += " "
	}
	text = strings.Replace(text, "\n", "\n"+iden, -1)
	return iden+text
}

func LastCommand() string {
	return strings.TrimSpace(lastCommand)
}

func LastOutput() string {
	return strings.TrimSpace(lastOutput)
}

func Command(expression string) bool {
	return regexp.MustCompile(expression).MatchString(lastCommand)
}

func OutputMatch(expression string) bool {
	return regexp.MustCompile(expression).MatchString(lastOutput)
}

func Output() {
	if !wasInteractive && !didOutput {
		fmt.Print(lastOutput)
		didOutput = true
	}
	lastOutputWasSay = false
}

func Prompt() bool {
	rows, columns := getsize()
	cmdline.Query(" stty rows " + strconv.Itoa(rows))
	cmdline.Query(" stty columns " + strconv.Itoa(columns))

	didOutput = false
	exec.Command("stty", "-F", "/dev/tty", "-echo", "-icanon", "min", "1").Run()
	defer exec.Command("stty", "-F", "/dev/tty", "sane").Run()
	var ok bool
	lastCommand, ok = cmdline.ReadCommand()
	if ! ok {
		// cli terminated
		return false
	}
	lastOutput, wasInteractive, ok = cmdline.ReadOutput()
	if ! ok {
		return false
	}
	Output()

	lastOutputWasSay = false
	return true
}

func Quit() {
	cmdline.Quit()
	running = false
}
