# Machine Problem 1 - Simple Shell

The work is a C++ implementation of a simple Unix-like shell. It supports running commands, input & output redirection, and pipeline.

## Problems

This work has to tackle four main challenges. First, parse the users' commands in regard to the shell syntax, which can be compound of commands, arguments, redirection symbols ("`>`" & "`<`"), and pipe symbol ("`|`"). Secondly, parsed commands in one user input should run in parallel as independent processes. Third, for redirections, the corresponding targets need to replace the `STDIN` or `STDOUT`. At last, pipe-linked processes require interprocess communication.

## Implementation

The main process is simply a never-ending while loop to catch the user input once at a time.

### Command Structure

A structure `CommandObj` is defined to store the necessary attributes of each executable command. The structure consists of a vector of tokens of execution command and arguments and two strings for redirection input and output file paths if any, and two integers working as booleans to indicate if the command has a pipe-in stream from the previous command and pipe-out stream to the next one. While these attributes can be extracted during parsing the users' input, it also contains a number for the process id which is not used in parsing stage but assigned values later in execution, which is mandatory for the main process to wait for the results of the children processes.

### Parsing

The user input is captured as a string. The program needs to parse the string into an array of the above defined command structure `CommandObj`. The input string is split into an array of string tokens with space. The parsing result is initialized as an array with only one empty command structure. The program iterates the string tokens. If the token is a redirection syntax symbol ("`>`" & "`<`"), the following string token is saved as the input or output file path in the command structure respectively. If the token is a pipe symbol ("`|`"), the pip out of the command structure is marked as true and a new command structure with pip in of true is pushed into the result array. Any following parsing operation then modifies this newly created structure instead until another pipe. For any other tokens, they simply pushed into the token vector of the command structure. In the end, the array with at least one command structure is formed to execute.

While doing parsing, the program also conducts syntax validation. Every command structure must have at least one token. There must be a non-syntax token following the redirection symbol. Each command structure can only have at most one of each redirection symbol. Any failure of these rules leads to "Invalid command" error.

### Execution

After parsing, the program iterates the parsed command objects to execute them in separate processes. If the command has the pipe-to marked as true (`1`), a pair of file descriptors are created and the write end is saved as the pipe to in the command structure. If the command has the pipe-from marked as true (`1`), the previous command must have also created a pair of file descriptors and the read end is saved as the pipe from in the command structure. Then the main process forks a child process and saves the process id in the command structure. Subsequently, the main process simply closes the write end of the newly created pipe file descriptor and the left read end of the previous pair if any, and go to the next iteration without waiting for this individual child process. After iterating all commands, the main process then iteratively wait for all the child processes to complete and go back to receive users input again. On the other hand, if the command has any pipe in or out, the child process duplicates them as the `STDIN` or `STDOUT` correspondingly and close the descriptors. Next, if the command has any redirections, the child process opens the file and duplicates them as the `STDIN` or `STDOUT` correspondingly. Then the child process replaces the current process image with a new process image by executing the array of tokens.

## Analysis

The work requires the user input to specify the executables in the absolute path because it does not have any `PATH`  variables. A command can have both the redirection and pipe which may cause conflicts, like `cat test.txt > copy.txt | cat`. There is no specification for how this should be handled. In this implementation, redirection has higher priority since the program duplicates the pipe to `STDIN/STDOUT` first then duplicates the redirections if any, which overwrites the previous duplication. For the input, this work does not help to do any formatting and naively split the input by space, so the user has to explicitly types the space to around the syntax special symbols ("`>`", "`<`", "`|`").

