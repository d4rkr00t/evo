echo "Devloop for EVO..."
watchexec -w ./ -e go "go build main.go  && cp ./main ~/.bin-temp/evo"
