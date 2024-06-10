# `dogz`

This tool was inspired by [a Stack Overflow question](https://stackoverflow.com/questions/78548374/extract-individual-files-from-concatenated-gzipped-files/78550323).

It exists to make it possible to extract individual members from a gzip stream with [multiple members](https://www.rfc-editor.org/rfc/rfc1952#page-5).

## Installation

```console
$ go install github.com/jonjohnsonjr/dogz@latest
```

## Usage

`dogz` will print the offset of the start of each gzip member in an archive.
[APKs](https://wiki.alpinelinux.org/wiki/Apk_spec) are one example of a multi-gzip member file format:

```console
$ curl -sL https://packages.wolfi.dev/os/aarch64/curl-8.8.0-r0.apk | dogz
0
706
1103
```

We can construct a multi-member gzip stream from scratch using standard tools:

```console
# Create some files to concatenate.
$ echo foo > foo.txt
$ echo bar > bar.txt
$ echo baz > baz.txt

# Create a multi-member gzip file.
$ cat <(gzip -c foo.txt) <(gzip -c bar.txt) <(gzip -c baz.txt) > catted.gz

# Verify that the concatenated gzip files decode correctly.
$ gunzip < catted.gz
foo
bar
baz
```

Notably, the `gzip` tool doesn't have a way to "undo" this operation.
This is what `dogz` is for.
It parses the stream and outputs the offset to the start of each member.

```console
$ dogz catted.gz
0 foo.txt
32 bar.txt
64 baz.txt

# We can use xxd to verify the gzip header is present at those offsets.
$ xxd -s0 -l3 catted.gz
00000000: 1f8b 08                                  ...
$ xxd -s32 -l3 catted.gz
00000020: 1f8b 08                                  ...
$ xxd -s64 -l3 catted.gz
00000040: 1f8b 08                                  ...

# If we wanted to skip the first gzip member.
$ tail -c +33 < catted.gz | gunzip
bar
baz

# Or extract only the second gzip member.
$ dogz tail -c +33 < catted.gz | head -c 32 | gunzip
bar

# Or extract only the final gzip member.
$ tail -c +65 < catted.gz | gunzip
baz
```

## Why `dogz`?

What's the opposite of `cat`?
