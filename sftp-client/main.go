package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kevinburke/ssh_config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

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
	var k string
	pubKey, err := cfg.Get(sshHost, "IdentityFile")
	if err != nil || len(pubKey) < 1 {
	  home := os.Getenv("HOME")
		k = fmt.Sprintf("%s/.ssh/id_rsa", home);
	} else {
		k = pubKey

	}
	private, err := getPrivateKeyPath(k)
	if err != nil {
		return nil, err
	}
  fmt.Sprintf("%d", private);

  var auths []ssh.AuthMethod

  if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
      auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
  }
  hostname, _ := cfg.Get(sshHost, "Hostname")

	return &ssh.ClientConfig{
		User: username,
		Auth: auths,
    HostKeyCallback: ssh.FixedHostKey(getHostKey(hostname)),
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
	hostname, err := config.Get(host, "Host")
	if err != nil {
		return "", errors.New(fmt.Sprintf("No host property for %s", host))
	}
	return fmt.Sprintf("%s:%s", hostname, port), nil
}

func exitProgram(err error) {
    fmt.Println("cnotes-sftp-client: Progam exited :(");
    fmt.Println(err);
    os.Exit(1);
}

func main() {
	var (
		err        error
		sftpClient *sftp.Client
	)

	filePtr := flag.String("file", "", "File to be uploaded")
	destPtr := flag.String("dest", "", "Destination folder")
	sshHostPtr := flag.String("ssh-host", "", "A host in the .ssh/config file")

	flag.Parse()

	if *filePtr == "" || *destPtr == "" {
		exitProgram(errors.New("Please provide file and destination arguments"))
	}
	cfg, err := loadSSHConfig()
	if err != nil {
		exitProgram(err)
	}
	hostPort, err := getHostAndPort(cfg, *sshHostPtr)
	if err != nil {
		exitProgram(err)
	}
  config, err := createSftpConfig(cfg, *sshHostPtr);
	if err != nil {
		exitProgram(err)
	}

	conn, err := ssh.Dial("tcp", hostPort, config)
	if err != nil {
		exitProgram(err)
	}
	defer conn.Close()

	sftpClient, err = sftp.NewClient(conn)
	if err != nil {
		exitProgram(err)
	}
	defer sftpClient.Close()

	srcFile, err := os.Open(*filePtr)
	if err != nil {
		exitProgram(err)
	}
	defer srcFile.Close()

	destFile := *destPtr + "/" + *filePtr
	stat, err := sftpClient.Stat(destFile)

	if os.IsNotExist(err) || stat.ModTime().Before(time.Now()) {
		dstFile, err := sftpClient.Create(destFile)
		if err != nil {
			exitProgram(err)
		}
		defer dstFile.Close()

    bytes, err := io.Copy(dstFile, srcFile)
		if err != nil {
			exitProgram(err)
		}

		fmt.Printf("%d bytes uploaded\n", bytes)
	} else {
		fmt.Println("The file is up-to-date.")
	}
}



// Get host key from local known hosts
func getHostKey(host string) ssh.PublicKey {
    // parse OpenSSH known_hosts file
    // ssh or use ssh-keyscan to get initial key
    file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to read known_hosts file: %v\n", err)
        os.Exit(1)
    }
    defer file.Close()
 
    scanner := bufio.NewScanner(file)
    var hostKey ssh.PublicKey
    for scanner.Scan() {
        fields := strings.Split(scanner.Text(), " ")
        if len(fields) != 3 {
            continue
        }
        if strings.Contains(fields[0], host) {
            var err error
            hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error parsing %q: %v\n", fields[2], err)
                os.Exit(1)
            }
            break
        }
    }
 
    if hostKey == nil {
        fmt.Fprintf(os.Stderr, "No hostkey found for %s", host)
        os.Exit(1)
    }
 
    return hostKey
}

func hashed() {

}
