# moby-demo

A demo application for Moby, the library that powers Docker.

```bash
$ make
mkdir dir1
mkdir dir2
go get -u
go build .
$ ./moby-demo
uname -a
(exited with 0)
eLinux 9e1114f658a9 6.3.5-arch1-1 #1 SMP PREEMPT_DYNAMIC Tue, 30 May 2023 13:44:01 +0000 x86_64 Linux
$ ./moby-demo whoami
whoami
(exited with 0)
root
$ make clean
rm --force go.sum moby-demo
rm --force --recursive dir1 dir2
```

This source code demonstrates pulling and removing images;
creating, starting, stopping, and interacting with containers;
and how common options are set with this API.

## License

I license this work under the BSD 3-clause license.

