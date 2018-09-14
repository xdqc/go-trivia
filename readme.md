# **Go-Quizzer** is an app to automatically solve quiz for you

## Used Technologies:

1. MITM server: goproxy

2. In-memory NoSQL database: boltDB

3. Image to text conversion

    a. OCR: tesseract, Google Vison API, Baidu OCR API
    
    b. Image hashing: dHash, aHash

4. Machine Learing for matching the closest option to quiz body, given google/wiki search result as reference.
    
    a. Sliding window sketch

    b. Kernel functions

    c. Weight matrix 
