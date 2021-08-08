echo "Devloop for SCU..."
watchexec -w ./ -e go "go build main.go && cp ./main ~/.bin-temp/scu"
