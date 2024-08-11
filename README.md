This is a Kenyan fleet/saccos management system 


## Table of Contents
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)
- [Known Issues](#known-issues)

## Prerequisites
- Go 1.21 or later
- MySQL 8.0

## Installation
1. Clone the repository:
    ```bash
    git clone https://github.com/MichaelWaruiru/SaccoManagement.git
    ```
2. Navigate to the project directory:
    ```bash
    cd ProjectDirectory
    ```
3. Install dependencies:
    ```bash
    go get ./...
    ```
4. Build the project:
    ```bash
    go build
    ```

## Usage
To start the web server:
 
    http://localhost:8080
    

### Contributing
If you want to contribute, these are the guidelines for doing so:

```markdown
1. Fork the repository.
2. Create a new branch (git checkout -b feature-branch main`).
3. Make your changes.
4. Commit your changes (`git commit -m 'Add some features/fix'`).
5. Push to the branch (`git push origin feature-branch main).
6. Open a pull request.

```
### License
MIT License

Copyright (c) [2024]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## Known Issues
- The API has issues in route and trips. A fix is in progress.
- Deleting sacco will cause errors in both the server and the user side.
