local:
	go run main.go -file=test.txt -dest=/ -ssh-hort=erk-temp

run-test-ssh-server:
	docker run --rm --name temp-fs -p 2222:2222 -e USER_NAME='erk' -e PUBLIC_KEY=(cat ~/.ssh/id_ed25519.pub) lscr.io/linuxserver/openssh-server:latest
