moby-demo: dir1 dir2
	go get -u
	go build .

.PHONY: clean
clean:
	rm --force go.sum moby-demo
	rm --force --recursive dir1 dir2

dir%:
	mkdir $@

