package database

import (
	"eman-backend/models"
	"log"
)

// SeedProjects populates projects table with default data if empty.
func SeedProjects() error {
	var count int64
	DB.Model(&models.Project{}).Count(&count)
	if count > 0 {
		// Update existing projects that have empty text fields
		var items []models.Project
		DB.Find(&items)
		defaults := defaultProjects()
		updated := 0
		for i, item := range items {
			if i >= len(defaults) {
				break
			}
			if item.TypeRu == "" || item.DescriptionRu == "" {
				d := defaults[i]
				item.TypeRu = d.TypeRu
				item.TypeUz = d.TypeUz
				item.AreaRu = d.AreaRu
				item.AreaUz = d.AreaUz
				item.DescriptionRu = d.DescriptionRu
				item.DescriptionUz = d.DescriptionUz
				if item.Image == "" {
					item.Image = d.Image
				}
				DB.Save(&item)
				updated++
			}
		}
		if updated > 0 {
			log.Printf("Updated %d projects with missing text fields", updated)
		}
		return nil
	}

	projects := defaultProjects()
	for _, p := range projects {
		if err := DB.Create(&p).Error; err != nil {
			log.Printf("Warning: Failed to seed project %s: %v", p.TypeRu, err)
		}
	}

	log.Printf("Seeded %d default projects", len(projects))
	return nil
}

func defaultProjects() []models.Project {
	return []models.Project{
		{
			TypeRu:        "Архитектура и материалы",
			TypeUz:        "Arxitektura va materiallar",
			AreaRu:        "Премиальное качество",
			AreaUz:        "Premium sifat",
			DescriptionRu: "<h3>ФАСАДЫ</h3><ul><li>Декоративная покраска</li><li>Фрезерованный металл</li><li>Металл под покраску</li><li>Природные оттенки</li><li>Современные формы</li></ul><h3>ОБЩИЕ ЗОНЫ</h3><ul><li>Керамогранит</li><li>Травертин</li><li>Износостойкие материалы</li></ul>",
			DescriptionUz: "<h3>FASADLAR</h3><ul><li>Dekorativ bo'yash</li><li>Frezerlangan metall</li><li>Bo'yash uchun metall</li><li>Tabiiy ranglar</li><li>Zamonaviy shakllar</li></ul><h3>UMUMIY ZONALAR</h3><ul><li>Keramogranit</li><li>Travertin</li><li>Bardoshli materiallar</li></ul>",
			Image:         "/images/hero/1.png",
			SortOrder:     1,
			IsPublished:   true,
		},
		{
			TypeRu:        "Студия",
			TypeUz:        "Studiya",
			AreaRu:        "от 28 м²",
			AreaUz:        "28 m² dan",
			DescriptionRu: "<p>Уютная студия с панорамными видами — идеальное пространство для молодых профессионалов и пар.</p><ul><li>Панорамные окна</li><li>Высокие потолки 3м</li><li>Функциональная планировка</li><li>Просторная кухня-гостиная</li></ul>",
			DescriptionUz: "<p>Panoramali ko'rinishga ega qulay studiya — yosh mutaxassislar va juftliklar uchun ideal makon.</p><ul><li>Panoramali derazalar</li><li>3m baland shiftlar</li><li>Funksional rejalashtirish</li><li>Keng oshxona-mehmonxona</li></ul>",
			Image:         "/images/hero/1.png",
			SortOrder:     2,
			IsPublished:   true,
		},
		{
			TypeRu:        "2-комнатная",
			TypeUz:        "2-xonali",
			AreaRu:        "от 55 м²",
			AreaUz:        "55 m² dan",
			DescriptionRu: "<p>Просторная двухкомнатная квартира с продуманной планировкой для комфортной семейной жизни.</p><ul><li>Раздельные комнаты</li><li>Мастер-спальня</li><li>Балкон с видом</li><li>Гардеробная</li></ul>",
			DescriptionUz: "<p>Qulay oilaviy hayot uchun o'ylangan rejalashtirishga ega keng ikki xonali kvartira.</p><ul><li>Alohida xonalar</li><li>Asosiy yotoqxona</li><li>Ko'rinishli balkon</li><li>Kiyim xonasi</li></ul>",
			Image:         "/images/hero/1.png",
			SortOrder:     3,
			IsPublished:   true,
		},
	}
}
