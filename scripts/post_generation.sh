cd proto/generated

cd ts
npm install
npm run build

cd ..
go mod tidy
