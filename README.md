Intro
=====

The diffprompt command reads a combination of prompt and code input, parses it, then uses ChatGPT to
operate on the input according to the instructions of the prompt. It is easy to integrate it with
your editor.

Setup
-----

Build and install the tool in your local golang bin directory:

```
rick@mac diffprompt % which diff
/usr/bin/diff
rick@mac diffprompt % echo $?
0
rick@mac diffprompt % go env|grep GOPATH
GOPATH='/Users/rick/go'
rick@mac diffprompt % pushd tools/diffprompt 
~/Documents/diffprompt/tools/diffprompt ~/Documents/diffprompt
rick@mac diffprompt % go build
rick@mac diffprompt % go install
rick@mac diffprompt % popd
~/Documents/diffprompt
rick@mac diffprompt % ls -lsa /Users/rick/go/bin/diffprompt
15312 -rwxr-xr-x  1 rick  staff  7838594 Feb 28 11:56 /Users/rick/go/bin/diffprompt
```

In the above commands, for a typical golang install, I first check that we have
the diff tool installed. The exit status should be 0, which it is. Next I check
where the tool will be installed by default, which is in the bin directory of
the GOPATH. I then build the tool, install it, and check that it is installed
in the expected location.

Next we need to store the OpenAI API key in a config file in our home directory:

```
echo "sk-proj-THIS_IS_MY_SECRET_API_KEY" > ~/.diffprompt
```

Now we are ready for an example run of the diffprompt tool:

```
#!/bin/bash

echo 'This is a test
vvv
func Test(t *testing.T) {

  // Implement test here for func Scan(array []int, value int) (int, bool) which returns first index with value and true, or false if none

}'|./diffprompt
```

Input to `diffprompt` is a combination of prompt and code. The prompt are all lines that come before
the delimeter line `vvv`. The code is all lines that come after the delimeter line.

Vim Integration
---------------
Add this line to your .vimrc file:

```
" Map g-c on visiual selection to run highlighted text through the diffprompt
" command
vnoremap gc :!diffprompt<CR>
```

Now you can highlight a block of text in visual mode and press `gc` to run it through diffprompt.

In the example below, we add the prompt, the delimeter line, then highlight the prompt and
code with Shift-V in `vim`, then press `gc` to run the code through diffprompt:

Before:
```
Modify sideBySideDiff so that it passes the input into diff as standard input
instead of as a temporary file.
vvv
// sideBySideDiff returns a side-by-side diff of the input and result strings by
// writing both strings to temporary files, then running `diff -y` on them.
func sideBySideDiff(input, result string) (string, error) {
	inputFile, err := os.CreateTemp("", "input-")
	if err != nil {
		return "", errors.Wrap(err, "creating input file failed")
	}
	defer os.Remove(inputFile.Name())

	resultFile, err := os.CreateTemp("", "result-")
	if err != nil {
		return "", errors.Wrap(err, "creating result file failed")
	}
	defer os.Remove(resultFile.Name())

	if _, err := inputFile.WriteString(input); err != nil {
		return "", errors.Wrap(err, "writing input file failed")
	}
	if _, err := resultFile.WriteString(result); err != nil {
		return "", errors.Wrap(err, "writing result file failed")
	}

	if err := inputFile.Close(); err != nil {
		return "", errors.Wrap(err, "closing input file failed")
	}
	if err := resultFile.Close(); err != nil {
		return "", errors.Wrap(err, "closing result file failed")
	}

	// diff --no-prefix -U1000
	cmd := exec.Command("diff", "-U10000000", inputFile.Name(), resultFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws, ok := exitError.Sys().(syscall.WaitStatus)
			if !ok {
				return "", errors.New("failed to get exit status")
			}

			exitCode := ws.ExitStatus()
			if exitCode > 1 {
				return "", errors.New("diff command failed")
			}
		} else {
			return "", errors.New("diff call failed")
		}
	}

	return strip3(string(output)), nil
}
```

After:
```
-// sideBySideDiff returns a side-by-side diff of the input and result strings by
-// writing both strings to temporary files, then running `diff -y` on them.
 func sideBySideDiff(input, result string) (string, error) {
-	inputFile, err := os.CreateTemp("", "input-")
-	if err != nil {
-		return "", errors.Wrap(err, "creating input file failed")
-	}
-	defer os.Remove(inputFile.Name())
+	cmd := exec.Command("diff", "-y", "-U10000000")
+	cmd.Stdin = strings.NewReader(input + "\n---\n" + result)
 
-	resultFile, err := os.CreateTemp("", "result-")
-	if err != nil {
-		return "", errors.Wrap(err, "creating result file failed")
-	}
-	defer os.Remove(resultFile.Name())
-
-	if _, err := inputFile.WriteString(input); err != nil {
-		return "", errors.Wrap(err, "writing input file failed")
-	}
-	if _, err := resultFile.WriteString(result); err != nil {
-		return "", errors.Wrap(err, "writing result file failed")
-	}
-
-	if err := inputFile.Close(); err != nil {
-		return "", errors.Wrap(err, "closing input file failed")
-	}
-	if err := resultFile.Close(); err != nil {
-		return "", errors.Wrap(err, "closing result file failed")
-	}
-
-	// diff --no-prefix -U1000
-	cmd := exec.Command("diff", "-U10000000", inputFile.Name(), resultFile.Name())
 	output, err := cmd.CombinedOutput()
 	if err != nil {
 		if exitError, ok := err.(*exec.ExitError); ok {
 			ws, ok := exitError.Sys().(syscall.WaitStatus)
 			if !ok {
 				return "", errors.New("failed to get exit status")
 			}
 
 			exitCode := ws.ExitStatus()
 			if exitCode > 1 {
 				return "", errors.New("diff command failed")
 			}
 		} else {
 			return "", errors.New("diff call failed")
 		}
 	}
 
 	return strip3(string(output)), nil
 }
```
