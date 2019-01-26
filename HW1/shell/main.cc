#include <cstdlib>
#include <iostream>
#include <string>
#include <sstream>
#include <vector>
#include <sys/types.h>
#include <unistd.h>
#include <sys/wait.h>

void parse_and_run_command(const std::string &command) {
    /* TODO: Implement this. */
    /* Note that this is not the correct way to test for the exit command.
       For example the command "   exit  " should also exit your shell.
    */

    std::istringstream ss (command);
    std::vector<std::string> tokens;
    std::string token;

    while (ss >> token) {
        tokens.push_back(token);
    }

    // std::cout << "received tokens: ";
    // for (auto i = tokens.begin(); i != tokens.end(); i++) {
    //     std::cout << *i << ", ";
    // }
    // std::cout << '\n';

    if (command == "exit") {
        exit(0);
    }

    pid_t pid = fork();

    if (pid < 0) {
        std::cerr << "fork command error";
        exit(1);
    }
    else if (pid == 0) {
        std::vector<char*> charTokens;
        for (auto i = tokens.begin(); i != tokens.end(); i++) {
            char* str = &(*i)[0];
            charTokens.push_back(str);
        }

        execvp(charTokens[0], charTokens.data());
    }
    else {
        wait(NULL);
    }

    // std::cerr << "Not implemented.\n";
}

int main(void) {
    while (true) {
        std::string command;
        std::cout << "> ";
        std::getline(std::cin, command);
        parse_and_run_command(command);
    }
    return 0;
}
