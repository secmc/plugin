cd proto/generated

cd ts
npm install
npm run build

cd ../go
rm -rf go.mod
go mod init github.com/secmc/plugin/proto/generated/go
go mod tidy
