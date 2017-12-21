package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func download_pipe(url string, filename string, dir string) error {
	cmd := exec.Command("wget", url, "-O", dir+"/"+filename)
	// cmd := exec.Command("aria2c", "-x", "8", "-c", url, "-o", dir+"/"+filename)
	return pipeCmd(cmd, os.Stdout)
}

func pipeCmd(cmd *exec.Cmd, w io.Writer) error {
	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdoutDone := make(chan struct{})
	scanner := bufio.NewScanner(stdoutReader)
	go func() {
		for scanner.Scan() {
			fmt.Fprint(w, scanner.Text())
		}
		stdoutDone <- struct{}{}
		logger.Printf("End of pipeCmd stdout goroutine: %v\n", scanner.Err())
	}()

	stderrDone := make(chan struct{})
	errScanner := bufio.NewScanner(stderrReader)
	go func() {
		for errScanner.Scan() {
			fmt.Fprintf(w, "\r")
			fmt.Fprint(w, errScanner.Text())
		}
		stderrDone <- struct{}{}
		fmt.Fprintln(w, "")
		logger.Printf("End of pipeCmd stderr goroutine: %v\n", errScanner.Err())
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}
	logger.Printf("pipeCmd start \n")

	<-stdoutDone
	<-stderrDone

	err = cmd.Wait()
	if err != nil {
		return err
	}

	logger.Println("end of pipeCmd.")
	return nil
}
