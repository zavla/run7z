// Creates 7z archives with special names needed for further maintanace.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
)

var excludePattern = flag.String("x", "*.$lk,*.mlg,*.cdx,*.LCK,*.log", "comma sepparated list of `excluded filename patterns`: ex.: *.$lk, *.cdx,")
var comment = flag.String("c", "1c77dir", "Additional `word` to the archive filename.")
var workingdir = flag.String("w", "", "`Where` create archive.")

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 ||
		*workingdir == "" {
		fmt.Println(`Using: run7z.exe [<switches>] -w <wherecreate> <dirnametoArchive>`)
		flag.PrintDefaults()
		log.Fatal("")
		return
	}

	dirnametoArchive := flag.Arg(0)
	filestat, err := os.Stat(dirnametoArchive)
	if err != nil {
		log.Fatalf("Dirname error: %s\n %v\n", dirnametoArchive, err)
	}
	if !filestat.IsDir() {
		log.Fatalf("Error: %s is not a dir.\n", dirnametoArchive)
	}

	rawexcludes := strings.Split(*excludePattern, ",")
	if len(rawexcludes) != 0 {
		for i, strval := range rawexcludes {
			rawexcludes[i] = "-xr!" + strings.TrimSpace(strval)
		}

	}

	const keyof7z = `SOFTWARE\7-Zip` //HKEY_LOCAL_MACHINE\SOFTWARE\7-Zip
	regkey, err := registry.OpenKey(registry.LOCAL_MACHINE, keyof7z, registry.READ)
	if err != nil {
		log.Fatalf(`Zaerror: Registry open for HKEY_LOCAL_MACHINE\%s failed.
		%s`, keyof7z, err.Error())
	}
	where7z, _, err := regkey.GetStringValue("Path64")
	if err != nil {
		log.Fatalf(`Zaerror: regkey.GetStringValue("Path64") failed
		%s`, err.Error())
	}
	path7z := filepath.Join(where7z, "7z.exe")
	fmt.Printf("running %s\n", path7z)

	//construct a filename needed
	additionaltextToFileName := strings.Replace(*comment, " ", "", -1)

	cleandirnametoArchive := filepath.Clean(dirnametoArchive)
	_, lastname := filepath.Split(cleandirnametoArchive)
	newArchiveFileName := fmt.Sprintf("%s-%s_%v", lastname, additionaltextToFileName, time.Now().Format("2006-01-02T15-04-05"))
	newArchiveFileNameFullPath := filepath.Join(*workingdir, newArchiveFileName)

	//wholeCmdline := fmt.Sprintf(" a -r %s %q %q", excludes, newArchiveFileNameFullPath, cleandirnametoArchive)
	//print(wholeCmdline)
	cmd := exec.Command(path7z)
	cmd.Args = append(cmd.Args, "a")
	cmd.Args = append(cmd.Args, "-r")
	cmd.Args = append(cmd.Args, "-ssw")
	for _, val := range rawexcludes {
		cmd.Args = append(cmd.Args, val)

	}
	cmd.Args = append(cmd.Args, "--")

	cmd.Args = append(cmd.Args, newArchiveFileNameFullPath)
	cmd.Args = append(cmd.Args, cleandirnametoArchive)

	errpipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Zaerror: cmd.StderrPipe failed.\n %s", err.Error())
	}
	if err = cmd.Start(); err != nil {
		log.Fatalf("Zaerror: start failed %s\n %s", path7z, err.Error())
	}
	outputbytesfromerr, err := ioutil.ReadAll(errpipe)
	if err != nil { // couldn't read stderr of the run process
		log.Fatalf("Zaerror: reading answer from running process failed.\n %s", err.Error())
	}
	err = cmd.Wait()
	if err != nil {
		// Archiver exited with non zero status
		fmt.Fprintf(os.Stderr, `Zaerror: Archiver exited with non zero status %s
		%s`, err.Error(), string(outputbytesfromerr))
	}
}
