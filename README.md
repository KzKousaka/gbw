Go Build Watch (gbw)
----------------------------------

This tools is command line tool.  
If save the source code, running for any command.

```
$gbw --command "echo write" --dir "./"
```

## Options

* **-command**  
  run command

* **-dir**  
  watching directory dirpath (default "./")

* **-ext**  
  extension list  (ex "png,gif")
  
* **-help**
  show help

* **-debug**  
  output debugging log

* **event options**  
  + created:  
    **-c**    enabled (default)  
    **-nc**   disabled  
  + writed:  
    **-w**    enabled (default)  
    **-nw**   disabled  
  + removed:  
    **-r**    enabled (default)  
    **-nr**   disabled  
  + renamed:  
    **-n**    enabled (default)  
    **-nn**   disabled  
  + change permission:  
    **-p**    enabled  
    **-np**   disabled (default)  

## Trable shoting

### error "Too meny open files"
The following is helpful  
EN: http://superuser.com/questions/830149/os-x-yosemite-too-many-files-open  
JP: http://ysh.hateblo.jp/entry/2015/11/26/211134

Fundamental revision requires refurbishment of [fsnotfy](https://github.com/fsnotify/fsnotify). (~_~)

## Update

* **2016.12.22**  
  Created ver 0.1