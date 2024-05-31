require 'sinatra'
require 'sequel'

DB = Sequel.connect('sqlite://fishbb.fb')

get '/' do erb :index end
