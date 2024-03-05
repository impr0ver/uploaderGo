# go-musthave-diploma-tpl
![Alt text](%D0%A2%D0%97_%D0%B2%D1%8B%D0%BF%D1%83%D1%81%D0%BA%D0%BD%D0%B0%D1%8F_%D1%80%D0%B0%D0%B1%D0%BE%D1%82%D0%B0_1.png)

«Система передачи данных на удаленный сервер или в «облачное» хранилище данных».

Список команд сервиса:

DTlQ1iuePre4AAAAAAAAAAdwpyxIBLy4hAYFPqYQsqNpitL2oMc_LH_Nzsh5tUvvT - токен который выдается при создании и настройке облачного хранилища DropBox. Токен нельзя терять, без него не получить временный токен доступа - access token!
# Список файлов в облаке DropBox
```
go run main.go -token=DTlQ1iuePre4AAAAAAAAAAdwpyxIBLy4hAYFPqYQsqNpitL2oMc_LH_Nzsh5tUvvT -listcloud

List folder status code: 200 OK
File name: myfile.txt, Cloud path: /data/myfile.txt, size: 18
File name: myfile3.txt, Cloud path: /data/myfile3.txt, size: 18
File name: myfile4.txt, Cloud path: /data/myfile4.txt, size: 18
File name: myfile5.txt, Cloud path: /data/myfile5.txt, size: 18
File name: myfile6.txt, Cloud path: /data/myfile6.txt, size: 18
File name: myfile7.txt, Cloud path: /data/myfile7.txt, size: 18
File name: myfile8.txt, Cloud path: /data/myfile8.txt, size: 18
```
# Удаление файла в облаке DropBox
```
go run main.go -token=DTlQ1iuePre4AAAAAAAAAAdwpyxIBLy4hAYFPqYQsqNpitL2oMc_LH_Nzsh5tUvvT -delete=/data/myfile8.txt

Delete file "/data/myfile8.txt" status code: 200 OK
```
# Загрузка файлов в облако DropBox
```
go run main.go -token=DTlQ1iuePre4AAAAAAAAAAdwpyxIBLy4hAYFPqYQsqNpitL2oMc_LH_Nzsh5tUvvT
или в режиме шифрования AES-256-CBC
go run main.go -token=DTlQ1iuePre4AAAAAAAAAAdwpyxIBLy4hAYFPqYQsqNpitL2oMc_LH_Nzsh5tUvvT -key=my_seceret

Work with cloud storage DropBox...
New accessToken: sl.Bw0Q2-J8VwDD2UA612wlvDbaEqhg1LQ37-54yLyrnP-hB6euRPDObRHDII2xccgycz39JL7G_ODEe3DWttx8Yw3kacyVWyrf...
Upload file "defaultfolder/TMP_3501760595/1709657123395352000_myfile.txt" status code: 200 OK
Upload file "defaultfolder/TMP_3501760595/1709657123396020000_myfile10.txt" status code: 200 OK
Upload file "defaultfolder/TMP_3501760595/1709657123396423000_myfile2.txt" status code: 200 OK
```
# Список файлов на файловом сервере
```
go run main.go -listserver
или в режиме шифрования AES-256-CBC (при условии что сервер тоже запущен в режиме шифрования - выполнено условие -key=)
go run main.go -listserver -key=my_secret

List folder status code: 200 OK
File name: uploads, Server path: ../../public/uploads, size: 160
File name: 1709645625856828000_myfile.txt, Server path: ../../public/uploads/1709645625856828000_myfile.txt, size: 18
File name: 1709645625857091000_myfile10.txt, Server path: ../../public/uploads/1709645625857091000_myfile10.txt, size: 18
```
# Удаление файла на файловом сервере
```
go run main.go -delete=1709645625856828000_myfi
Delete file "1709645625856828000_myfi" status code: 404 Not Found

go run main.go -delete=1709645625856828000_myfile.txt
Delete file "1709645625856828000_myfile.txt" status code: 200 OK
```
# Загрузка файлов на файловый сервер
```
режим multipart-form/data по 3 файла за раз (параметр 3 - по умолчанию)
go run main.go
или в режиме шифрования AES-256-CBC (если сервер тоже запущен в режиме шифрования с темже ключем то данные загрузятся в дешифрованном виде)
go run main.go -key=ggg

Sending files parts: [[defaultfolder/TMP_1606704178/1709659506579429000_myfile.txt defaultfolder/TMP_1606704178/1709659506580721000_myfile10.txt defaultfolder/TMP_1606704178/1709659506581575000_myfile2.txt] [defaultfolder/TMP_1606704178/1709659506582686000_myfile3.txt defaultfolder/TMP_1606704178/1709659506583503000_myfile4.txt defaultfolder/TMP_1606704178/1709659506583971000_myfile5.txt] [defaultfolder/TMP_1606704178/1709659506584656000_myfile6.txt defaultfolder/TMP_1606704178/1709659506585473000_myfile7.txt defaultfolder/TMP_1606704178/1709659506586273000_myfile8.txt] [defaultfolder/TMP_1606704178/1709659506586871000_myfile9.txt defaultfolder/TMP_1606704178/1709659506587489000_ТЗ_выпускная_работа_1_v4.pdf]]
Try to send part files: [defaultfolder/TMP_1606704178/1709659506579429000_myfile.txt defaultfolder/TMP_1606704178/1709659506580721000_myfile10.txt defaultfolder/TMP_1606704178/1709659506581575000_myfile2.txt]
Upload part files status code: 200 OK

или в режиме 1 to 1 через механизм горутин
go run main.go -mode=false -key=ggg

received: Upload file "defaultfolder/TMP_903909175/1709659753781919000_myfile4.txt" status code: 200 OK
received: Upload file "defaultfolder/TMP_903909175/1709659753780356000_myfile10.txt" status code: 200 OK
received: Upload file "defaultfolder/TMP_903909175/1709659753782283000_myfile5.txt" status code: 200 OK
received: Upload file "defaultfolder/TMP_903909175/1709659753784102000_myfile9.txt" status code: 200 OK
received: Upload file "defaultfolder/TMP_903909175/1709659753784551000_ТЗ_выпускная_работа_1_v4.pdf" status code: 200 OK
Wait for complete all routines...
```
# -h файлового сервера 
```
  -a string
    	Server address and port. (default "localhost:8443")
  -key string
    	Secret key for crypt/decrypt data with AES-256-CBC cipher algoritm.
  -workdir string
    	Path to store files. (default "/")
```
# -h файлового клиента
```
  -a string
        Server address and port. (default "localhost:8443")
  -delete string
        Delete file path.
  -fcount int
        Files count in one multipart/form-data body in POST request. (default 3)
  -key string
        Secret key for crypt data with AES-256-CBC cipher algoritm.
  -listcloud
        List data path (need set path).
  -listserver
        info
  -maxsize int
        Max size of send file in MB (<=16MB). (default 16)
  -mode
        Upload files mode: multipart (true) or single (false). (default true)
  -path string
        Path to find files for sending. (default "defaultfolder")
  -token string
        Refresh token for get access token and work with cloud storage.
```
