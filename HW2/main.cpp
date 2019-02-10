#include <string>
// #include <cstdlib>
#include <iostream>
#include <vector>


int main(void) {
    std::string input;
    std::vector<float> numbers;

    while (true) {
        std::getline(std::cin, input);

        if (input == "") {
            break;
        }

        try {
            float num = std::stof(input);
            numbers.push_back(num);
        } catch (std::exception& e) {
            std::cerr << "invalid input; exception: " << e.what() << '\n';
            continue;
        }
    }

    // for (auto f = numbers.begin(); f != numbers.end(); f++) {
    //     std::cout << *f << " ";
    // }
}
