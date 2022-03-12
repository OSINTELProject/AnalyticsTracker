#!/usr/bin/env python3
import sys
import re
import time
import yaml # pip install pyyaml
from pprint import pprint
import requests
from bs4 import BeautifulSoup

def read_yaml( file_path ):
	with open( file_path ) as f:
		return yaml.safe_load( f )

def convert_entry_to_dictionary( entry_items ):
	keys = [ "date" , "time" , "ip" , "country" , "lat_long" ]
	result = {}
	for index , item in enumerate( entry_items ):
		result[ keys[ index ] ] = item
	return result

def scrape_analytics_page( host , uuid ):
	try:
		url = f"{ host }/a/{ uuid }"
		response = requests.get( url )
		response.raise_for_status()
		html = response.text
		soup = BeautifulSoup( html , "html.parser" )

		total_views = soup.find_all( "h1" , string=re.compile( "^Total Views" ) )[ 0 ].text.split( " = " )[ 1 ]
		unique_views = soup.find_all( "h1" , string=re.compile( "^Unique Views" ) )[ 0 ].text.split( " = " )[ 1 ]

		entries = soup.find_all( "li" )
		parsed_entries = []
		for index , entry in enumerate( entries ):
			items = entry.text.split( " === " )
			if items[ -1 ] == "Map":
				items.pop()
			entry_dictionary = convert_entry_to_dictionary( items )
			parsed_entries.append( entry_dictionary )
		return {
			"total_views": total_views ,
			"unique_views": unique_views ,
			"entries": parsed_entries
		}
	except Exception as e:
		print( e )
		return False

if __name__ == "__main__":
	tracking = read_yaml( sys.argv[ 1 ] )
	parsed_analytics = []
	for index , name in enumerate( tracking[ "uuids" ] ):
		analytics = scrape_analytics_page( tracking[ "host" ] , tracking[ "uuids" ][ name ] )
		parsed_analytics.append( analytics )
		time.sleep( 1 )
	pprint( parsed_analytics )