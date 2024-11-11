# exttlinter

A Go linter that checks for use of external test object.

```go
func Test_a(t *testing.T) {
	assert := func() {
        // When used in subtests, it can cause issues with test failure reporting locations being different than expected.
		t.Error("no match") // Detecting the use of external test objects
	}

	t.Run("sub test", func(t *testing.T) {
		assert()
	})
}
```

## Usage

```sh
go vet -vettool=$(which exttlinter) ./...
```

```yml
# Github Actions workflow

- name: Set up Go
  uses: actions/setup-go@v2

- name: install vettool
  run : GOBIN=$(pwd) go install github.com/tomat3713/exttlinter@latest

- name: run vet
  run: go vet -vettool=$(pwd)/exttlinter ./...
```
