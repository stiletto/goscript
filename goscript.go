package main

import (
	"io"
    "syscall"
    "fmt"
    "os"
    "bufio"
    "crypto/md5"
    "encoding/hex"
    "net/textproto"
)

func stage_check(cpid int,err os.Error,stage string) {
    if err == nil {
        lolf, _ := os.Wait(cpid,0);
        if lolf.ExitStatus() != 0 {
            fmt.Printf("Failed to %s.\n",stage)
            os.Exit(lolf.ExitStatus())
        }
    } else {
        fmt.Println(err)
        os.Exit(1)
    }
}

var GOBIN = "/opt/go/bin/"
func main() {
    if len(os.Args) < 2 {
        fmt.Printf("Usage: %s <script file>\n",os.Args[0])
        return
    }
    namehash_hash := md5.New()
    io.WriteString(namehash_hash,os.Args[1])
    namehash := hex.EncodeToString(namehash_hash.Sum())
    tempfile := os.TempDir() + "/" + namehash;

    readpipe,writepipe, _ := os.Pipe()
    stdfiles := [](*os.File) {readpipe,os.Stdout,os.Stderr}

    syscall.Umask(0077)
    cpid, cerr := os.ForkExec(GOBIN+"8g",[]string{GOBIN+"8g","-o",tempfile+".cd","/dev/stdin"},nil,"",stdfiles)

    srcfile, err := os.Open(os.Args[1],os.O_RDONLY,0)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    src := textproto.NewReader(bufio.NewReader(srcfile))
    firstline := true
    for {
        s, err := src.ReadLine()
        if err != nil {
            break
        }
        if (len(s) > 0) && (s[0]=='#') && firstline {
            continue
        }
        firstline = false
        writepipe.WriteString(s+"\n")
    }
    writepipe.Close()

    stage_check(cpid,cerr,"compile")

    lpid, err := os.ForkExec(GOBIN+"8l",[]string{GOBIN+"8l","-o",tempfile+".ld",tempfile+".cd"},nil,"",stdfiles)
    stage_check(lpid,err,"link")
    os.Chmod(tempfile+".ld",0700) // i wish i have umode
    err = os.Exec(tempfile+".ld",os.Args[1:],nil)
    fmt.Println(err)
    os.Exit(1)
}
