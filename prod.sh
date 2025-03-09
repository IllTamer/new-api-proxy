cd web/
npm run build
cd ../
go build -ldflags "-s -w" -o new-api
sudo systemctl restart one-api