package app

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// LnxExecutePyScriptSsh connects to a remote Linux machine via SSH,
// transfers a local script to the remote machine using SFTP, and then executes it.
//
// Parameters:
//
//	host: The IP address or hostname of the remote Linux machine.
//	port: The SSH port of the remote machine (e.g., "22").
//	user: The username for SSH authentication on the remote machine.
//	localScriptPath: The local path to the script file to be transferred and executed.
//
// Note:
//   - This function currently uses hardcoded paths for the SSH private key
//     ("/path/to/private/key") and the remote script path ("/path/to/remote/script.sh").
//     These should be made configurable or passed as parameters for production use.
//   - It uses `ssh.InsecureIgnoreHostKey()` which is highly insecure and
//     should NOT be used in production environments. For secure connections,
//     proper host key verification should be implemented.
//   - Error handling uses `log.Fatalf` which will terminate the program on error.
//     Consider returning errors instead for more robust error management.
func LnxExecutePyScriptSsh(host string, port string, user string, localScriptPath string) {
	// Connection parameters
	// IP address or hostname of the remote machine
	if host == "" {
		host = "127.0.0.1"
	}
	// SSH port, usually 22
	if port == "" {
		port = "22"

	}
	// Username for the remote machine
	if user == "" {
		user = "demo"
	}
	privateKeyPath := "/path/to/private/key"        // Path to the SSH private key
	remoteScriptPath := "/path/to/remote/script.sh" // Remote path to store the script

	// Load the private key
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key: %v", err)
	}

	// Create an SSH key signer
	keySigner, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// SSH client configuration
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(keySigner),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Insecure host key checking
	}

	// Connect to the SSH server
	address := fmt.Sprintf("%s:%s", host, port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer client.Close()

	// Open an SFTP session
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		log.Fatalf("Failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	// Open the local script
	localFile, err := os.Open(localScriptPath)
	if err != nil {
		log.Fatalf("Failed to open local file: %v", err)
	}
	defer localFile.Close()

	// Create the remote file
	remoteFile, err := sftpClient.Create(remoteScriptPath)
	if err != nil {
		log.Fatalf("Failed to create remote file: %v", err)
	}
	defer remoteFile.Close()

	// Copy the local script to the remote server
	_, err = localFile.Seek(0, 0) // Ensure reading from the start
	if err != nil {
		log.Fatalf("Failed to seek local file: %v", err)
	}

	_, err = localFile.WriteTo(remoteFile)
	if err != nil {
		log.Fatalf("Failed to write file to remote server: %v", err)
	}

	fmt.Println("Script transferred successfully!")

	// Execute the script on the remote server
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Command to make the script executable and run it
	cmd := fmt.Sprintf("chmod +x %s && %s", remoteScriptPath, remoteScriptPath)
	err = session.Run(cmd)
	if err != nil {
		log.Fatalf("Failed to run command: %v", err)
	}

	fmt.Println("Script executed successfully on remote machine.")
}
