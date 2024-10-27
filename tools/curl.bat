remove-item alias:curl
curl -X POST http://func-1.application.127.0.0.1.sslip.io -H "Content-Type: image/jpeg"  -H "X-Forward-To: imagegrab" --data-binary "@image.txt"