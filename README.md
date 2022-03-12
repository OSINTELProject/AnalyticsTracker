# Analytics Tracking Server

## Setup

1. `cp config_example.json config.json`
2. `nano config.json`
3. `./dockerBuild.sh`
4. `./dockerRun.sh`
5. Configure Caddy Reverse Proxy :
```
ta.example.org {
        reverse_proxy localhost:9337 {
                header_up Host {upstream_hostport}
                header_up X-Forwarded-Host {host}
        }
}
```

## Example Usage
- Generate Some UUID `uuid`
- Add to some HTML
- `<img src='https://ta.example.com/t/$SOME_UUID' alt=''>`
- or
```
try {
	let img_1 = document.createElement( "img" );
	img_1.setAttribute( "src" , "https://ta.example.com/t/$SOME_UUID?v=" + ( new Date() ).getTime() );
	img_1.style.display = "none";
	document.body.appendChild( img_1 );
} catch( e ) {}
```
- View Results : https://ta.example.com/a/$SOME_UUID
