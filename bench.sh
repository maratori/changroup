echo "Install bench stat"
go install golang.org/x/perf/cmd/benchstat@latest
echo ""
echo "Bench container/list"
go test -bench=. -benchmem -count 10 | tee std.txt
echo ""
echo "Bench custom generic list"
go test -bench=. -benchmem -count 10 -short | tee generic.txt
echo ""
echo "Compare"
benchstat std.txt generic.txt
