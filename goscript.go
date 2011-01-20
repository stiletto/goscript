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
func temp_dir() string {
    dirname := os.TempDir() + "/goscript." + fmt.Sprintf("%d",os.Geteuid())
    err := os.Mkdir(dirname,0700)
    if err != nil {
        err := err.(*os.PathError)
        if err.Error == os.EEXIST {
            dirfile, err := os.Open(dirname,os.O_RDONLY,0)
            if (err != nil) {
                fmt.Println(err)
                os.Exit(1)
            }
            info, _ := dirfile.Stat()
            if (info.Uid != os.Geteuid()) || ((info.Mode&0777) != 0700) {
                fmt.Println("Temporary directory "+dirname+" has wrong ownership or permissions.")
                os.Exit(1)
            }
        } else {
            fmt.Println(err)
            os.Exit(1)
        }
    }
    return dirname+"/"
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
    tempfile := temp_dir() + "/" + namehash;

    readpipe,writepipe, _ := os.Pipe()
    stdfiles := [](*os.File) {readpipe,os.Stdout,os.Stderr}

    syscall.Umask(0077)
    srcfile, err := os.Open(os.Args[1],os.O_RDONLY,0)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    srcinfo, _ := srcfile.Stat()
    outfile, err := os.Open(tempfile+".ld",os.O_RDONLY,0)
    var outinfo (*os.FileInfo)
    if err == nil {
        outinfo, _ = outfile.Stat()
        outfile.Close()
    }

    if err != nil || outinfo.Mtime_ns < srcinfo.Mtime_ns {
        //fmt.Printf("Compiling...")
        cpid, cerr := os.ForkExec(GOBIN+"8g",[]string{GOBIN+"8g","-o",tempfile+".cd","/dev/stdin"},nil,"",stdfiles)

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
        srcfile.Close()

        stage_check(cpid,cerr,"compile")

        lpid, err := os.ForkExec(GOBIN+"8l",[]string{GOBIN+"8l","-o",tempfile+".ld",tempfile+".cd"},nil,"",stdfiles)
        stage_check(lpid,err,"link")
        os.Chmod(tempfile+".ld",0700)
        fmt.Println("done.")
    } else {
        //fmt.Println("Running cached.")
    }
    err = os.Exec(tempfile+".ld",os.Args[1:],nil)
    fmt.Println(err)
    os.Exit(1)
}
