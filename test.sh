if [ "$1" == "-v" ]; then
  go test ./lib -v
elif [ "$1" == "-c" ]; then
  go test ./lib -coverprofile cov.html && go tool cover -html=cov.html
else
  go test ./lib $@
fi
