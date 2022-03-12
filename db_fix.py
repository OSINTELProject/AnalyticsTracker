#!/usr/bin/env python3
import redis
import json

def write_json( file_path , python_object ):
	with open( file_path , 'w', encoding='utf-8' ) as f:
		json.dump( python_object , f , ensure_ascii=False , indent=4 )

def read_json( file_path ):
	with open( file_path ) as f:
		return json.load( f )


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
	for index , record in enumerate( records ):
		if "74.140.161.64" not in record:
			print( f"{index} === {record}" )
			cleaned_records.append( record )
	redis_connection.delete( "ANALYTICS.e34d9b65-79ba-4109-aa21-904bbc6d3c68.RECORDS" )
	for index , record in enumerate( cleaned_records ):
		redis_connection.rpush( "ANALYTICS.e34d9b65-79ba-4109-aa21-904bbc6d3c68.RECORDS" , record )