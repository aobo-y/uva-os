# Simple Shell

A Unix-like shell that supports running commands as described below, that includes:

* running simple commands (e.g. /bin/cat foo.txt bar.txt)
* input redirection (e.g. commands like /usr/bin/gcc -E - < somefile.txt)
* output redirection (e.g. commands like /usr/bin/gcc -E file.cc > somefile.txt)
* supports the builtin command exit In addition, your shell must
* not search for executables on PATH; all commands will include the absolute path to the executable or a relative path from the current directory
* prompt for commands using >
* print out error messages (to stderr):
  * including the text “Invalid command” if a command is malformed according to the language specification below;
  * including either the text “Command not found” or “No such file or directory” if the executable does not exist;
  * including some error message (we do not care which one) for errors when fork, pipe, exec, or opening a file for redirection fails.
* implement pipelines (e.g. commands like /bin/cat foo.txt | /bin/grep bar | /bin/grep baz)
* output to stdout the exit status(es) of every command using lines containing “exit status: “ followed by the numerical exit status of each command (if the command terminates normally; if the command exits from a signal, we do not care what is output). For a pipeline that includes multiple commands, output exit statuses for the parts of the pipeline in the order they appear in the command.
* not leave any extra file descriptors open (ones other than stdin, stdout, stderr) just before it calls executes a command.


## Specification

### Shell language

Shell commands are lines which consist of a sequence of whitespace-seperated tokens. Whitespace characters are space (' ' in C or C++), form feed ('\f'), newline ('\n'), carriage return ('\r'), horizontal tab ('\t'), vertical tab ('\v').

Each line has a maximum length of 100 characters. We do not care what your shell does with longer input lines.

Each token is either

* an operator, if it is < or > or |, or
* a word otherwise

Each line of inputs is a pipeline, which consists of a series of a commands (possibly just one) seperated by |s.

Each command consists of a series of:

* one or more words, which are used to form the command for an exec system call
* up to one input redirection operation, which consists of a < token followed by a word token
* up to one output redirection operation, which consists of a > token followed by a word token

### Running commands

If a command contains excess input or output redirection operations, or operators followed immediately by other operator, that is an error.

To run a pipeline, the shell runs each command in the pipeline, then waits for all the commands to terminate. You may decide what to do if one of the commands in the pipeline cannot be executed as long as your shell does not crash or otherwise stop accepting new commands.

To run a command, the shell:

* first checks if it is one of the built-in commands listed below, and if so does something special.
* forks off a subprocess for the command
* if it is not the first command in the pipeline, connect its stdin (file descriptor 0) to the stdout of the previous command in the pipeline. The connection should be created as if by pipe (see man pipe). (You may not, for example, create a normal temporary file, even if this sometimes works.)
* if it is not the last command in the pipeline, connect its stdout (file descriptor 1) to the stdin of the next command in the pipeline
* if there is an output redirection operation, reopen stdout (file descriptor 1) to that file. The file should be created if it does not exist, and truncated if it does.
* if there is an input redirection operation, reopen stdin (file descriptor 0) from that file
* uses the first word as the name of the executable run.

### Built-in command

Your shell should support the built-in command exit. When this command is run your shell should terminate normally (with exit status 0). You should not run a program called exit, even if one exists.

### Handling errors

Your shell should print out error messages to stderr (file descriptor 2).

You must use the following error messages:

* If an executable does not exist, print an error message containing “Command not found” or “No such file or directory”. You may include other text in the error message (perhaps the name of the executable the user was trying to run). Note that “No such file or directory” is what perror or strerror will output for an errno value of ENOENT. (See their manpages by running man perror or man strerror.)
* If a command is malformed according to the language above, print an error message containing “Invalid command”

You must also print out an error message if:

* exec fails for another reason (e.g. executable found but not executable)
* fork or pipe fail for any reason. Your program must not crash in this case.
* opening a input or output redirected file fails for any reason.

If multiple commands in a pipeline fail, you must print out error messages, but it may be the case that commands in the pipeline are executed even though error messages are printed out (e.g. it’s okay if something_that_does_not_exist | something_real prints an error about something_that_does_not_exist after starting something_real.)
