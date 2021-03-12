del /Q nekoq-security
del /Q nekoq-security_linux.tgz
set GOOS=linux
go build
tar czf nekoq-security_linux.tgz nekoq-security nekoq-security.toml.example
del /Q nekoq-security
