@echo off
cd /d E:\Projects\projectx\AllOpsFull
go build ./internal/common/ 1>NUL 2>_buildcheck.txt && echo common: PASS >> _buildcheck.txt || echo common: FAIL >> _buildcheck.txt
go build ./internal/app/ 1>NUL 2>_builderr.txt && echo app: PASS >> _buildcheck.txt || (echo app: FAIL >> _buildcheck.txt & type _builderr.txt >> _buildcheck.txt)
go build ./internal/netops/ 1>NUL 2>NUL && echo netops: PASS >> _buildcheck.txt || echo netops: FAIL >> _buildcheck.txt
go build ./internal/secops/ 1>NUL 2>NUL && echo secops: PASS >> _buildcheck.txt || echo secops: FAIL >> _buildcheck.txt
go build ./internal/devops/ 1>NUL 2>NUL && echo devops: PASS >> _buildcheck.txt || echo devops: FAIL >> _buildcheck.txt
go vet ./internal/networkdesign/ 1>NUL 2>NUL && echo netdesign-vet: PASS >> _buildcheck.txt || echo netdesign-vet: FAIL >> _buildcheck.txt
go test ./internal/networkdesign/ -count=1 1>NUL 2>NUL && echo netdesign-test: PASS >> _buildcheck.txt || echo netdesign-test: FAIL >> _buildcheck.txt
echo DONE >> _buildcheck.txt
