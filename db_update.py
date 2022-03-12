#!/usr/bin/env python3
import json
import redis
import requests
import time
from pprint import pprint

def write_json( file_path , python_object ):
	with open( file_path , 'w', encoding='utf-8' ) as f:
		json.dump( python_object , f , ensure_ascii=False , indent=4 )

def read_json( file_path ):
	with open( file_path ) as f:
		return json.load( f )

def get_geo_info_for_ip( ip_address , access_token ):
	try:
		url = f"https://ipinfo.io/{ip_address}?token={access_token}"
		response = requests.get( url )
		response.raise_for_status()
		return response.json()
	except Exception as e:
		print( e )
		return False

# https://redis-py.readthedocs.io/en/stable/index.html?highlight=StrictRedis#redis.StrictRedis
if __name__ == "__main__":
	config = read_json( "./remote.json" )
	redis_connection = redis.StrictRedis(
		host=config[ "redis" ][ "host" ] ,
		port=config[ "redis" ][ "port" ] ,
		db=config[ "redis" ][ "db" ] ,
		password=config[ "redis" ][ "password" ] ,
		decode_responses=True
	)
	records = redis_connection.lrange( "ANALYTICS.e34d9b65-79ba-4109-aa21-904bbc6d3c68.RECORDS" , 0 , -1 )
	cleaned_records = []
	unique = {}
	for index , record in enumerate( records ):
		ip = record.split( " === " )[ -1 ]
		if "Map" in ip:
			cleaned_records.append( record )
			continue
		if ip not in unique:
			print( ip )
			unique[ ip ] = 1
			geo_info = get_geo_info_for_ip( ip , config[ "ip_info" ][ "token" ] )
			if "loc" in geo_info:
				map_url = f'<a href="https://maps.google.com/?q={geo_info["loc"]}">Map</a>'
				record = f"{record} === {geo_info[ 'country' ]} === {geo_info['loc']} === {map_url}"
				cleaned_records.append( record )
			else:
				cleaned_records.append( record )
			time.sleep( 1 )
		else:
			cleaned_records.append( record )
			continue
	redis_connection.delete( "ANALYTICS.e34d9b65-79ba-4109-aa21-904bbc6d3c68.RECORDS" )
	for index , record in enumerate( cleaned_records ):
		redis_connection.rpush( "ANALYTICS.e34d9b65-79ba-4109-aa21-904bbc6d3c68.RECORDS" , record )
	pprint( cleaned_records )


	# redis_connection.delete( "ANALYTICS.e34d9b65-79ba-4109-aa21-904bbc6d3c68.RECORDS" )
	# for index , record in enumerate( cleaned_records ):
	# 	redis_connection.rpush( "ANALYTICS.e34d9b65-79ba-4109-aa21-904bbc6d3c68.RECORDS" , record )