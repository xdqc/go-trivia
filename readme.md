# **Go-Quizzer** is an app to automatically solve quiz for you

## Technologies that I played around in this side project:

1. **MITM** server: goproxy

2. In-memory **NoSQL** database: boltDB

3. Image to text conversion

    a. **OCR**: tesseract, Google Vison API, Baidu OCR API
    
    b. Image **hashing**: dHash, aHash

4. **Machine Learing** for matching the closest option to quiz body, given google/wiki search result as reference.
    
    a. Sliding window sketch methodology

    b. SVM and kernal tricks
    
        - Gaussian kernal
        - Exponential kernel

    c. Weight matrix 
    
    d. Covariance matrix

All of these implemented in ```Golang``` from scratch, which gave me significantly higher quizzing accuracy than the naive frequency-counting approach used by most relevant projects. 
