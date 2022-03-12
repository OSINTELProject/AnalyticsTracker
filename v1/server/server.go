package server

import (
	"os"
	"fmt"
	"time"
	"net"
	"strings"
	"context"
	"reflect"
	fiber "github.com/gofiber/fiber/v2"
	rate_limiter "github.com/gofiber/fiber/v2/middleware/limiter"
	redis_manager "github.com/0187773933/RedisManagerUtils/manager"
	try "github.com/manucorporat/try"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

type IPInfoResponse struct {
	Bogon bool `json:bogon`
	IP string `json:ip`
	City string `json:city`
	Region string `json:region`
	Country string `json:country`
	Location string `json:loc`
	Organization string `json:org`
	Postal string `json:postal`
	TimeZone string `json:timezone`
}

func GetIPGeoInfo( ip_address string , ip_info_token string ) ( result map[string]interface{} ) {
	try.This( func() {
		url := fmt.Sprintf( "https://ipinfo.io/%s?token=" , ip_address , ip_info_token )
		resp, err := http.Get( url )
		if err != nil { fmt.Println( err ) }
		defer resp.Body.Close()
		body , err := ioutil.ReadAll( resp.Body )
		if err != nil { fmt.Println( err ) }
		body_string := string( body )
		fmt.Println( body_string )

		// V1.) Anonymous JSON Decode
		json_data_reader := strings.NewReader( body_string )
		json_decoder := json.NewDecoder( json_data_reader )
		json_decoding_error := json_decoder.Decode( &result )
		if json_decoding_error != nil { fmt.Println( json_decoding_error ) }
		fmt.Println( result )

		// V2.) Typed JSON Decode
		// json_decode_error := json.Unmarshal( body , &result )
		// if json_decode_error != nil { fmt.Println( json_decode_error ); return }
		// fmt.Println( result )

		// V3.) Other
		// json_data_reader := strings.NewReader( body_string )
		// json_decoder := json.NewDecoder( json_data_reader )
		// json_decoding_error := json_decoder.Decode( &result )
		// if json_decoding_error != nil { fmt.Println( json_decoding_error ) }
		// fmt.Println( result )

		// if val , ok := dict["foo"]; ok {
		// 	//do something here
		// }
	}).Catch( func( e try.E ) {
		fmt.Printf( "failed to lookup geoinfo for : %s\n" , ip_address )
	})
	return
}

func GetRedisConnection( host string , port string , db int , password string ) ( redis_client redis_manager.Manager ) {
	redis_client.Connect( fmt.Sprintf( "%s:%s" , host , port ) , db , password )
	return
}

func RedisSetAdd( redis redis_manager.Manager , set_key string , value string ) ( unique bool ) {
	unique = false
	var ctx = context.Background()
	set_add_result , set_add_error := redis.Redis.SAdd( ctx , set_key , value ).Result()
	if set_add_error != nil { fmt.Println( set_add_error ); }
	if set_add_result == 1 { unique = true }
	return
}

func RedisSetGetSize( redis redis_manager.Manager , set_key string ) ( set_card_result int64 ) {
	var ctx = context.Background()
	set_card_result , set_card_error := redis.Redis.SCard( ctx , set_key ).Result()
	if set_card_error != nil { fmt.Println( set_card_error ); }
	return
}

func RedisGetList( redis redis_manager.Manager , list_key string ) ( result []string ) {
	var ctx = context.Background()
	list_result , list_get_error := redis.Redis.LRange( ctx , list_key , 0 , -1 ).Result()
	if list_get_error != nil { fmt.Println( list_get_error ); }
	result = list_result
	return
}


// https://stackoverflow.com/a/28862477
func GetLocalIPAddresses() ( ip_addresses []string ) {
	host , _ := os.Hostname()
	addrs , _ := net.LookupIP( host )
	for _ , addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			// fmt.Println( "IPv4: " , ipv4 )
			ip_addresses = append( ip_addresses , ipv4.String() )
		}
	}
	return
}

func GetFormattedTimeString( time_zone string ) ( result string ) {
	location , _ := time.LoadLocation( time_zone )
	time_object := time.Now().In( location )
	// https://stackoverflow.com/a/51915792
	// month_name := strings.ToUpper( time_object.Format( "Feb" ) ) // monkaHmm
	month_name := strings.ToUpper( time_object.Format( "Jan" ) )
	milliseconds := time_object.Format( ".000" )
	date_part := fmt.Sprintf( "%02d%s%d" , time_object.Day() , month_name , time_object.Year() )
	time_part := fmt.Sprintf( "%02d:%02d:%02d%s" , time_object.Hour() , time_object.Minute() , time_object.Second() , milliseconds )
	result = fmt.Sprintf( "%s === %s" , date_part , time_part )
	return
}

type RedisConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	DB int `json:"db"`
	Password string `json:"password"`
}
type ServerConfig struct {
	TimeZone string `json:"time_zone"`
	Redis RedisConfig `json:"redis"`
	IPBlacklist []string `json:"ip_blacklist"`
	IPInfoToken string `json:"ip_info_token"`
}
func New( config *ServerConfig ) ( app *fiber.App ) {
	app = fiber.New()
	ip_addresses := GetLocalIPAddresses()
	fmt.Println( ip_addresses )
	fmt.Println( reflect.TypeOf( app ) )
	// https://docs.gofiber.io/api/middleware/limiter
	app.Use( rate_limiter.New( rate_limiter.Config{
		Max: 2 ,
		Expiration: ( 4 * time.Second ) ,
		// Next: func( c *fiber.Ctx ) bool {
		// 	ip := c.IP()
		// 	fmt.Println( ip )
		// 	return ip == "127.0.0.1"
		// } ,
		LimiterMiddleware: rate_limiter.SlidingWindow{} ,
		KeyGenerator: func( c *fiber.Ctx ) string {
			return c.Get( "x-forwarded-for" )
		} ,
		LimitReached: func( c *fiber.Ctx ) error {
			ip := c.IP()
			other_ips := c.IPs()
			if len( other_ips ) > 0 {
				ip = other_ips[ 0 ]
			}
			fmt.Printf( "%s === limit reached\n" , ip )
			c.Set( "Content-Type" , "text/html" )
			return c.SendString( "<html><h1>why</h1></html>" )
		} ,
		// Storage: myCustomStorage{}
		// monkaS
		// https://github.com/gofiber/fiber/blob/master/middleware/limiter/config.go#L53
	}))
	// tracking := app.Group( "/t" )
	app.Get( "/t/:id" , func( fiber_context *fiber.Ctx ) ( error ) {
		id := fiber_context.Params( "id" )
		ip := fiber_context.IP()
		other_ips := fiber_context.IPs()
		if len( other_ips ) > 0 {
			ip = other_ips[ 0 ]
		}
		for _ , v := range config.IPBlacklist {
			if v == ip {
				// "shadow" ignore
				fiber_context.Set( fiber.HeaderContentType , fiber.MIMETextHTML )
				return fiber_context.SendString( "<html><h1>new tracking</h1></html>" )
			}
		}
		try.This( func() {
			global_total_key := fmt.Sprintf( "ANALYTICS.%s.TOTAL" , id )
			// global_unique_total_key := fmt.Sprintf( "ANALYTICS.%s.UNIQUE_TOTAL" , id )
			global_ips_key := fmt.Sprintf( "ANALYTICS.%s.IPS" , id )
			global_records_key := fmt.Sprintf( "ANALYTICS.%s.RECORDS" , id )
			// ip_total_key := fmt.Sprintf( "ANALYTICS.%s.%s.TOTAL" , id , ip )
			// ip_times_key := fmt.Sprintf( "ANALYTICS.%s.%s.TIMES" , id , ip )

			time_string := GetFormattedTimeString( config.TimeZone )

			redis := GetRedisConnection( config.Redis.Host , config.Redis.Port , config.Redis.DB , config.Redis.Password )
			redis.Increment( global_total_key )
			// redis.Increment( ip_total_key )
			// redis.ListPushRight( ip_times_key , time_string )

			// Build and Store Record
			record := fmt.Sprintf( "%s === %s" , time_string , ip )
			unique := RedisSetAdd( redis , global_ips_key , ip )
			if unique == true {
				ip_info := GetIPGeoInfo( ip , config.IPInfoToken )
				fmt.Println( ip_info )
				if val , ok := ip_info[ "country" ]; ok {
					record = fmt.Sprintf( "%s === %s" , record , val )
				}
				if val , ok := ip_info[ "org" ]; ok {
					record = fmt.Sprintf( "%s === %s" , record , val )
				}
				if val , ok := ip_info[ "loc" ]; ok {
					google_maps_url := fmt.Sprintf( "<a href=\"https://maps.google.com/?q=%s\">Map</a>" , val )
					record = fmt.Sprintf( "%s === %s === %s" , record , val , google_maps_url )
				}
			}
			fmt.Println( record )
			redis.ListPushRight( global_records_key , record )
		}).Catch( func( e try.E ) {
			fmt.Printf( "failed to register new tracking info for : %s\n" , id )
		})
		fiber_context.Set( fiber.HeaderContentType , fiber.MIMETextHTML )
		return fiber_context.SendString( "<html><h1>new tracking</h1></html>" )
	})
	app.Get( "/a/:id" , func( fiber_context *fiber.Ctx ) ( error ) {
		id := fiber_context.Params( "id" )
		ip := fiber_context.IP()
		other_ips := fiber_context.IPs()
		if len( other_ips ) > 0 {
			ip = other_ips[ 0 ]
		}
		var html_result_string string
		try.This( func() {
			global_total_key := fmt.Sprintf( "ANALYTICS.%s.TOTAL" , id )
			// global_unique_total_key := fmt.Sprintf( "ANALYTICS.%s.UNIQUE_TOTAL" , id )
			global_ips_key := fmt.Sprintf( "ANALYTICS.%s.IPS" , id )
			global_records_key := fmt.Sprintf( "ANALYTICS.%s.RECORDS" , id )

			redis := GetRedisConnection( config.Redis.Host , config.Redis.Port , config.Redis.DB , config.Redis.Password )
			total_views := redis.Get( global_total_key )
			total_unique_views := RedisSetGetSize( redis , global_ips_key )
			fmt.Printf( "Total Unique Views === %v\n" , total_unique_views )

			records := RedisGetList( redis , global_records_key )
			records_html_string := "<ol>\n"
			for _ , v := range records {
				records_html_string = records_html_string + fmt.Sprintf( "<li>%v</li>\n" , v )
			}
			records_html_string = records_html_string + "</ol>"
			// fmt.Println( records )

			fmt.Printf( "Total Views === %v\n" , total_views )
			fmt.Printf( "%s === analytics\n" , ip )
			html_result_string = fmt.Sprintf( "<html>\n\t<h1>Total Views = %s</h1>\n<h1>Unique Views = %d</h1>\n%s</html>" , total_views , total_unique_views , records_html_string )
		}).Catch( func( e try.E ) {
			fmt.Printf( "failed to lookup analytics info for : %s\n" , id )
		})
		fiber_context.Set( fiber.HeaderContentType , fiber.MIMETextHTML )
		return fiber_context.SendString( html_result_string )
	})
	return
}