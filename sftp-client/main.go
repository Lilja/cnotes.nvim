package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strings"

	"github.com/kevinburke/ssh_config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)
var version = "1"

var known_hosts_file = fmt.Sprintf("%s/.ssh/known_hosts", os.Getenv("HOME"))

func loadSSHConfig() (*ssh_config.Config, error) {
	home := os.Getenv("HOME")
	if len(home) < 1 {
		return nil, errors.New("No HOME variable")
	}
	dat, err := os.ReadFile(fmt.Sprintf("%s/.ssh/config", home))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("No %s/.ssh/config", home))
	}
	return ssh_config.Decode(strings.NewReader(string(dat)))
}

func createSftpConfig(
	cfg *ssh_config.Config,
	sshHost string,
) (*ssh.ClientConfig, error) {
	username, err := cfg.Get(sshHost, "User")
	if err != nil {
		return nil, err
	}

	hkCallback, err := knownhosts.New(known_hosts_file)
	if err != nil {
		log.Println("No known host file?")
		return nil, err
	}

	var auths []ssh.AuthMethod

	aconn, sockErr := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))

	if sockErr == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
	}

	fmt.Println(len(auths))

	return &ssh.ClientConfig{
		User:            username,
		Auth:            auths,
		HostKeyCallback: hkCallback,
	}, nil
}

func getPrivateKeyPath(privateKey string) (ssh.Signer, error) {
	fmt.Println(privateKey)
	privateBytes, err := os.ReadFile(privateKey)
	if err != nil {

		fmt.Println("asdf")
		return nil, err
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil, err
	}
	return private, nil
}

func getHostAndPort(
	config *ssh_config.Config,
	host string,
) (string, error) {
	var port string
	_port, err := config.Get(host, "Port")
	if err != nil {
		port = "22"
	} else {
		port = _port
	}
	hostname, err := config.Get(host, "Hostname")
	if err != nil {
		return "", errors.New(fmt.Sprintf("No host property for %s", host))
	}
	return fmt.Sprintf("%s:%s", hostname, port), nil
}

func exitProgram(err error) {
	fmt.Println("cnotes-sftp-client: Progam exited :(")
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	var (
		err        error
		sftpClient *sftp.Client
	)

	filePtr := flag.String("file", "", "File to be uploaded")
	destPtr := flag.String("dest", "", "Destination folder")
	sshHostPtr := flag.String("ssh-host", "", "A host in the .ssh/config file")
	forceUpload := flag.Bool("force", false, "Force upload if file already exists")
	log.Println(fmt.Sprint("Version ", version))

	flag.Parse()

	localFile := *filePtr
	if localFile == "" || *destPtr == "" {
		exitProgram(errors.New("Please provide file and destination arguments"))
	}
	if strings.Contains(*filePtr, "/") {
		log.Println("Cleaning file name")
		localFile = path.Base(*filePtr)
	}
	cfg, err := loadSSHConfig()
	if err != nil {
		log.Println("load ssh config")
		exitProgram(err)
	}
	hostPort, err := getHostAndPort(cfg, *sshHostPtr)
	if err != nil {
		log.Println("get hsot and port")
		exitProgram(err)
	}
	config, err := createSftpConfig(cfg, *sshHostPtr)
	if err != nil {
		log.Println("create sftp config")
		exitProgram(err)
	}

	conn, err := ssh.Dial("tcp", hostPort, config)
	if err != nil {
		log.Println("Error dialing up tcp connection")
		if strings.Contains(err.Error(), "knownhosts: key is unknown") {
			exitProgram(errors.New(fmt.Sprintf("No known host entry for this host. Please try to connect to it manually and/or set the known host entry in %s", known_hosts_file)))
		} else {
			exitProgram(err)
		}
	}
	defer conn.Close()

	sftpClient, err = sftp.NewClient(conn)
	if err != nil {
		log.Println("sftp client error")
		exitProgram(err)
	}
	defer sftpClient.Close()

	srcFile, err := os.Open(*filePtr)
	if err != nil {
		log.Println("os open error")
		exitProgram(err)
	}
	defer srcFile.Close()
	srcFileStat, err := os.Stat(*filePtr)
	if err != nil {
		log.Println("os stat error")
		exitProgram(err)
	}

	destFile := *destPtr + "/" + localFile
	stat, err := sftpClient.Stat(destFile)

	if os.IsNotExist(err) {
		uploadFile(sftpClient, destFile, srcFile)
	} else {
		obuf := bytes.NewBufferString("")
		rd := bufio.NewWriter(obuf)
		remoteFile, err := sftpClient.Open(destFile)
		if err != nil {
			exitProgram(errors.New("Unable to check checksumz"))
		}
		downloadFileToInMemory(rd, remoteFile)
		rd.Flush()

		var flag = true
		if compareVersions(obuf, *filePtr) {
			fmt.Println("OK: Up to date!")
			flag = false
		} else if stat.ModTime().Before(srcFileStat.ModTime()) {
			if *forceUpload == false {
				fmt.Println("File exists on remote server and is created earlier than current file. Pass -force to forcefully upload file")
			}
		} else {
			if *forceUpload == false {
				fmt.Println("File exists on remote server and is created later than current file. Pass -force to forcefully upload file")
			}
		}
		if *forceUpload && flag {
			uploadFile(sftpClient, destFile, srcFile)
		}
	}
}

func uploadFile(sftpClient *sftp.Client, destFile string, srcFile *os.File) {
	log.Println(destFile, "will be created on the remote")
	dstFile, err := sftpClient.Create(destFile)
	if err != nil {
		log.Println("Unable to create a file")
		exitProgram(err)
	}
	defer dstFile.Close()

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		log.Println("Error copying file to remote place")
		exitProgram(err)
	}
	fmt.Printf("OK: %d bytes uploaded\n", bytes)
}

func downloadFileToInMemory(inMemoryWriter io.Writer, srcFile *sftp.File) {
	bytes, err := io.Copy(inMemoryWriter, srcFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to download remote file: %v\n", err)
		os.Exit(1)
	}
	log.Println(bytes, "bytes read")
}

func compareVersions(
	remoteFile *bytes.Buffer,
	fileName string,
) bool {
	remoteFileSha := sha256.New()
	remoteFileSha.Write(remoteFile.Bytes())

	localFileBuffer := bytes.NewBufferString("")
	rd := bufio.NewWriter(localFileBuffer)

	localFile, err := os.Open(fileName)
	if err != nil {
		log.Println("os open error")
		exitProgram(err)
	}
	defer localFile.Close()

	io.Copy(rd, localFile)
	rd.Flush()
	localFileSha := sha256.New()
	localFileSha.Write(localFileBuffer.Bytes())

	lb := localFileSha.Sum(nil)
	rb := remoteFileSha.Sum(nil)
	return bytes.Equal(lb, rb)
}
