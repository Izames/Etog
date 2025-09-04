package main

import "Etog/internal/config"

func main() {
	conf := config.MustLoad()
	println(conf)
	//TODO: разработать базу данных при помощи миграций
	//TODO: сделать обращения к базе данных
	//TODO: создать запросы для апи
	//TODO: создать бизнес-логику для выполнения АПИ
	//TODO: связать все
	//TODO: изучить тестирование и пройтись им по всему проекту
}
