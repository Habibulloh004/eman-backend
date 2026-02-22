package models

import (
	"time"
)

// SiteSetting represents a key-value configuration setting
type SiteSetting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"uniqueIndex;size:100" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	Type      string    `json:"type"`     // string, json, number, boolean
	Category  string    `json:"category"` // contact, social, pricing, faq, features, content
	Label     string    `json:"label"`    // Admin display label
	LabelUz   string    `json:"label_uz"` // Uzbek label
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SettingCategory constants
const (
	CategoryContact  = "contact"
	CategorySocial   = "social"
	CategoryPricing  = "pricing"
	CategoryFAQ      = "faq"
	CategoryFeatures = "features"
	CategoryContent  = "content"
	CategoryProjects = "projects"
	CategoryGallery  = "gallery"
)

// SettingType constants
const (
	TypeString  = "string"
	TypeJSON    = "json"
	TypeNumber  = "number"
	TypeBoolean = "boolean"
)

// DefaultSettings returns the default site settings for seeding
func DefaultSettings() []SiteSetting {
	return []SiteSetting{
		// Contact settings
		{Key: "phone", Value: "+998 90 123 45 67", Type: TypeString, Category: CategoryContact, Label: "Телефон", LabelUz: "Telefon"},
		{Key: "email", Value: "info@eman-riverside.uz", Type: TypeString, Category: CategoryContact, Label: "Email", LabelUz: "Email"},
		{Key: "address", Value: "Город Ташкент, Улица Богишев, Дом М71", Type: TypeString, Category: CategoryContact, Label: "Адрес", LabelUz: "Manzil"},
		{Key: "address_uz", Value: "Toshkent shahri, Botkin ko'chasi, 1-uy", Type: TypeString, Category: CategoryContact, Label: "Адрес (UZ)", LabelUz: "Manzil (UZ)"},
		{Key: "working_hours", Value: "Пн-Пт: 9:00 - 18:00", Type: TypeString, Category: CategoryContact, Label: "Время работы", LabelUz: "Ish vaqti"},
		{Key: "working_hours_uz", Value: "Du-Ju: 9:00 - 18:00", Type: TypeString, Category: CategoryContact, Label: "Время работы (UZ)", LabelUz: "Ish vaqti (UZ)"},

		// Social media settings
		{Key: "telegram", Value: "https://t.me/emanriverside", Type: TypeString, Category: CategorySocial, Label: "Telegram", LabelUz: "Telegram"},
		{Key: "instagram", Value: "https://instagram.com/emanriverside", Type: TypeString, Category: CategorySocial, Label: "Instagram", LabelUz: "Instagram"},
		{Key: "facebook", Value: "https://facebook.com/emanriverside", Type: TypeString, Category: CategorySocial, Label: "Facebook", LabelUz: "Facebook"},
		{Key: "youtube", Value: "https://youtube.com/@emanriverside", Type: TypeString, Category: CategorySocial, Label: "YouTube", LabelUz: "YouTube"},
		{Key: "threads", Value: "https://www.threads.net/@emanriverside", Type: TypeString, Category: CategorySocial, Label: "Threads", LabelUz: "Threads"},
		{Key: "whatsapp", Value: "+998901234567", Type: TypeString, Category: CategorySocial, Label: "WhatsApp", LabelUz: "WhatsApp"},

		// Pricing settings (JSON)
		{Key: "payment_plans", Value: `[
			{
				"title": "Ипотека",
				"description": "Удобное финансирование для покупки квартиры вашей мечты",
				"price": "1 млн сум",
				"period": "В месяц",
				"features": ["Первоначальный взнос от 30%", "Срок до 36 месяцев", "Без процентов"]
			},
			{
				"title": "Рассрочка",
				"description": "Гибкие условия рассрочки без дополнительных платежей",
				"price": "2 млн сум",
				"period": "В месяц",
				"features": ["Первоначальный взнос от 20%", "Срок до 24 месяцев", "Скидка 5%"]
			}
		]`, Type: TypeJSON, Category: CategoryPricing, Label: "Планы оплаты", LabelUz: "To'lov rejalari"},
		{Key: "payment_plans_uz", Value: `[
			{
				"title": "Ipoteka",
				"description": "Orzu qilgan kvartirangizni sotib olish uchun qulay moliyalashtirish",
				"price": "1 mln so'm",
				"period": "Oyiga",
				"features": ["Boshlang'ich to'lov 30% dan", "Muddat 36 oygacha", "Foizsiz"]
			},
			{
				"title": "Bo'lib to'lash",
				"description": "Qo'shimcha to'lovsiz moslashuvchan bo'lib to'lash shartlari",
				"price": "2 mln so'm",
				"period": "Oyiga",
				"features": ["Boshlang'ich to'lov 20% dan", "Muddat 24 oygacha", "5% chegirma"]
			}
		]`, Type: TypeJSON, Category: CategoryPricing, Label: "Планы оплаты (UZ)", LabelUz: "To'lov rejalari (UZ)"},

		// FAQ settings (JSON)
		{Key: "faq_items", Value: `[
			{"question": "Какие документы нужны для покупки?", "answer": "Для покупки квартиры вам понадобится паспорт и ИНН. Для оформления ипотеки дополнительно потребуется справка о доходах."},
			{"question": "Можно ли посмотреть квартиру?", "answer": "Да, вы можете записаться на просмотр квартиры, связавшись с нашим менеджером."},
			{"question": "Какие способы оплаты доступны?", "answer": "Мы предлагаем наличный расчет, банковский перевод, ипотеку и рассрочку."},
			{"question": "Когда будет сдан дом?", "answer": "Актуальные сроки сдачи уточняйте у наших менеджеров."},
			{"question": "Есть ли парковка?", "answer": "Да, в комплексе предусмотрены подземный и наземный паркинги."}
		]`, Type: TypeJSON, Category: CategoryFAQ, Label: "FAQ", LabelUz: "FAQ"},
		{Key: "faq_items_uz", Value: `[
			{"question": "Sotib olish uchun qanday hujjatlar kerak?", "answer": "Kvartira sotib olish uchun sizga pasport va STIR kerak bo'ladi. Ipoteka rasmiylashtirish uchun qo'shimcha ravishda daromad haqida ma'lumotnoma talab qilinadi."},
			{"question": "Kvartirani ko'rish mumkinmi?", "answer": "Ha, menejerimiz bilan bog'lanib kvartirani ko'rishga yozilishingiz mumkin."},
			{"question": "Qanday to'lov usullari mavjud?", "answer": "Biz naqd pul, bank o'tkazmasi, ipoteka va bo'lib to'lash taklif qilamiz."},
			{"question": "Uy qachon topshiriladi?", "answer": "Joriy topshirish muddatlarini menejerlarimizdan aniqlang."},
			{"question": "Avtoturargoh bormi?", "answer": "Ha, majmuada yer osti va yer usti avtoturargohlar mavjud."}
		]`, Type: TypeJSON, Category: CategoryFAQ, Label: "FAQ (UZ)", LabelUz: "FAQ (UZ)"},

		// Content settings
		{Key: "hero_title", Value: "EMAN RIVERSIDE", Type: TypeString, Category: CategoryContent, Label: "Заголовок Hero", LabelUz: "Hero sarlavha"},
		{Key: "hero_subtitle", Value: "Жилой комплекс нового уровня", Type: TypeString, Category: CategoryContent, Label: "Подзаголовок Hero", LabelUz: "Hero tavsif"},
		{Key: "hero_subtitle_uz", Value: "Yangi darajadagi turar-joy majmuasi", Type: TypeString, Category: CategoryContent, Label: "Подзаголовок Hero (UZ)", LabelUz: "Hero tavsif (UZ)"},
		{Key: "map_embed_url", Value: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d47980.98675893856!2d69.21992457431642!3d41.31147339999999!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x38ae8b0cc379e9c3%3A0xa5a9323b4aa5cb98!2sTashkent%2C%20Uzbekistan!5e0!3m2!1sen!2s!4v1703955000000!5m2!1sen!2s", Type: TypeString, Category: CategoryContent, Label: "Google Maps URL", LabelUz: "Google Maps URL"},
		{Key: "map_coordinates", Value: "41.3111,69.2401", Type: TypeString, Category: CategoryContent, Label: "Координаты карты", LabelUz: "Xarita koordinatalari"},
		{Key: "map_zoom", Value: "14", Type: TypeNumber, Category: CategoryContent, Label: "Масштаб карты", LabelUz: "Xarita masshtabi"},
		{Key: "background_music_url", Value: "", Type: TypeString, Category: CategoryContent, Label: "Фоновая музыка (URL)", LabelUz: "Fon musiqa URL"},
		{Key: "brochure_file_url", Value: "", Type: TypeString, Category: CategoryContent, Label: "Файл буклета", LabelUz: "Buklet fayli"},

		// Projects/Features settings (JSON) - 3 default features
		{Key: "projects", Value: `[
			{
				"number": "01",
				"label": "Современный Дизайн",
				"title": "Комфортное",
				"titleLine2": "жилье",
				"image": "/images/hero/1.png",
				"items": [
					{"title": "Фасады", "list": ["Керамогранит", "Алюминиевые панели", "Панорамное остекление"]},
					{"title": "Общие зоны", "list": ["Просторные холлы", "Современные лифты", "Подземный паркинг"]}
				]
			},
			{
				"number": "02",
				"label": "Благоустройство",
				"title": "Территория",
				"titleLine2": "для жизни",
				"image": "/images/hero/1.png",
				"items": [
					{"title": "Озеленение", "description": "Профессиональный ландшафтный дизайн с использованием взрослых деревьев"},
					{"title": "Детская площадка", "description": "Безопасные игровые зоны с современным оборудованием"},
					{"title": "Спортивная зона", "description": "Открытые площадки для активного отдыха"}
				]
			},
			{
				"number": "03",
				"label": "Инженерные системы",
				"title": "Надежные",
				"titleLine2": "коммуникации",
				"image": "/images/hero/1.png",
				"description": "Современные инженерные решения обеспечивают комфорт и безопасность",
				"features": ["Центральное кондиционирование", "Приточно-вытяжная вентиляция", "Система умный дом", "Видеонаблюдение"]
			}
		]`, Type: TypeJSON, Category: CategoryProjects, Label: "Проекты/Особенности", LabelUz: "Loyihalar/Xususiyatlar"},
		{Key: "projects_uz", Value: `[
			{
				"number": "01",
				"label": "Zamonaviy Dizayn",
				"title": "Qulay",
				"titleLine2": "turar-joy",
				"image": "/images/hero/1.png",
				"items": [
					{"title": "Fasadlar", "list": ["Keramogranit", "Alyuminiy panellar", "Panoramali oynalar"]},
					{"title": "Umumiy zonalar", "list": ["Keng zallar", "Zamonaviy liftlar", "Yer osti avtoturargoh"]}
				]
			},
			{
				"number": "02",
				"label": "Obodonlashtirish",
				"title": "Hayot uchun",
				"titleLine2": "hudud",
				"image": "/images/hero/1.png",
				"items": [
					{"title": "Ko'kalamzorlashtirish", "description": "Katta daraxtlar bilan professional landshaft dizayni"},
					{"title": "Bolalar maydoni", "description": "Zamonaviy jihozlar bilan xavfsiz o'yin zonalari"},
					{"title": "Sport zonasi", "description": "Faol dam olish uchun ochiq maydonchalar"}
				]
			},
			{
				"number": "03",
				"label": "Muhandislik tizimlari",
				"title": "Ishonchli",
				"titleLine2": "kommunikatsiyalar",
				"image": "/images/hero/1.png",
				"description": "Zamonaviy muhandislik yechimlari qulaylik va xavfsizlikni ta'minlaydi",
				"features": ["Markaziy konditsioner", "Kirish-chiqish ventilyatsiyasi", "Aqlli uy tizimi", "Videokuzatuv"]
			}
		]`, Type: TypeJSON, Category: CategoryProjects, Label: "Проекты/Особенности (UZ)", LabelUz: "Loyihalar/Xususiyatlar (UZ)"},

		// Gallery settings (JSON)
		{Key: "gallery_items", Value: `[
			{"image": "/images/hero/1.png", "title": "Фасад здания", "description": "Современный дизайн фасада с использованием премиальных материалов"},
			{"image": "/images/hero/1.png", "title": "Входная группа", "description": "Просторный холл с дизайнерской отделкой"},
			{"image": "/images/hero/1.png", "title": "Территория", "description": "Благоустроенная территория с зонами отдыха"},
			{"image": "/images/hero/1.png", "title": "Детская площадка", "description": "Безопасная игровая зона для детей"},
			{"image": "/images/hero/1.png", "title": "Паркинг", "description": "Подземный паркинг с видеонаблюдением"}
		]`, Type: TypeJSON, Category: CategoryGallery, Label: "Галерея", LabelUz: "Galereya"},
		{Key: "gallery_items_uz", Value: `[
			{"image": "/images/hero/1.png", "title": "Bino fasadi", "description": "Premium materiallar bilan zamonaviy fasad dizayni"},
			{"image": "/images/hero/1.png", "title": "Kirish guruhi", "description": "Dizayner pardozli keng zal"},
			{"image": "/images/hero/1.png", "title": "Hudud", "description": "Dam olish zonalari bilan obodonlashtirilgan hudud"},
			{"image": "/images/hero/1.png", "title": "Bolalar maydoni", "description": "Bolalar uchun xavfsiz o'yin zonasi"},
			{"image": "/images/hero/1.png", "title": "Avtoturargoh", "description": "Videokuzatuvli yer osti avtoturargoh"}
		]`, Type: TypeJSON, Category: CategoryGallery, Label: "Галерея (UZ)", LabelUz: "Galereya (UZ)"},
	}
}
