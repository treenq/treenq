package tqsdk

type SizeSlug string

const (

	// Ram: 0.5GB, Cpu: 1, Disk: 10GB, Transfer: 0.5TB, Price: 4$
	SizeSlugs1vcpu512mb10gb SizeSlug = "s-1vcpu-512mb-10gb"

	// Ram: 1GB, Cpu: 1, Disk: 25GB, Transfer: 1TB, Price: 6$
	SizeSlugs1vcpu1gb SizeSlug = "s-1vcpu-1gb"

	// Ram: 1GB, Cpu: 1, Disk: 25GB, Transfer: 1TB, Price: 7$
	SizeSlugs1vcpu1gbAmd SizeSlug = "s-1vcpu-1gb-amd"

	// Ram: 1GB, Cpu: 1, Disk: 25GB, Transfer: 1TB, Price: 7$
	SizeSlugs1vcpu1gbIntel SizeSlug = "s-1vcpu-1gb-intel"

	// Ram: 1GB, Cpu: 1, Disk: 35GB, Transfer: 1TB, Price: 8$
	SizeSlugs1vcpu1gb35gbIntel SizeSlug = "s-1vcpu-1gb-35gb-intel"

	// Ram: 2GB, Cpu: 1, Disk: 50GB, Transfer: 2TB, Price: 12$
	SizeSlugs1vcpu2gb SizeSlug = "s-1vcpu-2gb"

	// Ram: 2GB, Cpu: 1, Disk: 50GB, Transfer: 2TB, Price: 14$
	SizeSlugs1vcpu2gbAmd SizeSlug = "s-1vcpu-2gb-amd"

	// Ram: 2GB, Cpu: 1, Disk: 50GB, Transfer: 2TB, Price: 14$
	SizeSlugs1vcpu2gbIntel SizeSlug = "s-1vcpu-2gb-intel"

	// Ram: 2GB, Cpu: 1, Disk: 70GB, Transfer: 2TB, Price: 16$
	SizeSlugs1vcpu2gb70gbIntel SizeSlug = "s-1vcpu-2gb-70gb-intel"

	// Ram: 2GB, Cpu: 2, Disk: 60GB, Transfer: 3TB, Price: 18$
	SizeSlugs2vcpu2gb SizeSlug = "s-2vcpu-2gb"

	// Ram: 2GB, Cpu: 2, Disk: 60GB, Transfer: 3TB, Price: 21$
	SizeSlugs2vcpu2gbAmd SizeSlug = "s-2vcpu-2gb-amd"

	// Ram: 2GB, Cpu: 2, Disk: 60GB, Transfer: 3TB, Price: 21$
	SizeSlugs2vcpu2gbIntel SizeSlug = "s-2vcpu-2gb-intel"

	// Ram: 2GB, Cpu: 2, Disk: 90GB, Transfer: 3TB, Price: 24$
	SizeSlugs2vcpu2gb90gbIntel SizeSlug = "s-2vcpu-2gb-90gb-intel"

	// Ram: 4GB, Cpu: 2, Disk: 80GB, Transfer: 4TB, Price: 24$
	SizeSlugs2vcpu4gb SizeSlug = "s-2vcpu-4gb"

	// Ram: 4GB, Cpu: 2, Disk: 80GB, Transfer: 4TB, Price: 28$
	SizeSlugs2vcpu4gbAmd SizeSlug = "s-2vcpu-4gb-amd"

	// Ram: 4GB, Cpu: 2, Disk: 80GB, Transfer: 4TB, Price: 28$
	SizeSlugs2vcpu4gbIntel SizeSlug = "s-2vcpu-4gb-intel"

	// Ram: 4GB, Cpu: 2, Disk: 120GB, Transfer: 4TB, Price: 32$
	SizeSlugs2vcpu4gb120gbIntel SizeSlug = "s-2vcpu-4gb-120gb-intel"

	// Ram: 8GB, Cpu: 2, Disk: 100GB, Transfer: 5TB, Price: 42$
	SizeSlugs2vcpu8gbAmd SizeSlug = "s-2vcpu-8gb-amd"

	// Ram: 4GB, Cpu: 2, Disk: 25GB, Transfer: 4TB, Price: 42$
	SizeSlugc2 SizeSlug = "c-2"

	// Ram: 4GB, Cpu: 2, Disk: 50GB, Transfer: 4TB, Price: 47$
	SizeSlugc22vcpu4gb SizeSlug = "c2-2vcpu-4gb"

	// Ram: 8GB, Cpu: 2, Disk: 160GB, Transfer: 5TB, Price: 48$
	SizeSlugs2vcpu8gb160gbIntel SizeSlug = "s-2vcpu-8gb-160gb-intel"

	// Ram: 8GB, Cpu: 4, Disk: 160GB, Transfer: 5TB, Price: 48$
	SizeSlugs4vcpu8gb SizeSlug = "s-4vcpu-8gb"

	// Ram: 8GB, Cpu: 4, Disk: 160GB, Transfer: 5TB, Price: 56$
	SizeSlugs4vcpu8gbAmd SizeSlug = "s-4vcpu-8gb-amd"

	// Ram: 8GB, Cpu: 4, Disk: 160GB, Transfer: 5TB, Price: 56$
	SizeSlugs4vcpu8gbIntel SizeSlug = "s-4vcpu-8gb-intel"

	// Ram: 8GB, Cpu: 2, Disk: 25GB, Transfer: 4TB, Price: 63$
	SizeSlugg2vcpu8gb SizeSlug = "g-2vcpu-8gb"

	// Ram: 8GB, Cpu: 4, Disk: 240GB, Transfer: 6TB, Price: 64$
	SizeSlugs4vcpu8gb240gbIntel SizeSlug = "s-4vcpu-8gb-240gb-intel"

	// Ram: 8GB, Cpu: 2, Disk: 50GB, Transfer: 4TB, Price: 68$
	SizeSluggd2vcpu8gb SizeSlug = "gd-2vcpu-8gb"

	// Ram: 8GB, Cpu: 2, Disk: 30GB, Transfer: 4TB, Price: 76$
	SizeSlugg2vcpu8gbIntel SizeSlug = "g-2vcpu-8gb-intel"

	// Ram: 8GB, Cpu: 2, Disk: 60GB, Transfer: 4TB, Price: 79$
	SizeSluggd2vcpu8gbIntel SizeSlug = "gd-2vcpu-8gb-intel"

	// Ram: 16GB, Cpu: 4, Disk: 200GB, Transfer: 8TB, Price: 84$
	SizeSlugs4vcpu16gbAmd SizeSlug = "s-4vcpu-16gb-amd"

	// Ram: 16GB, Cpu: 2, Disk: 50GB, Transfer: 4TB, Price: 84$
	SizeSlugm2vcpu16gb SizeSlug = "m-2vcpu-16gb"

	// Ram: 8GB, Cpu: 4, Disk: 50GB, Transfer: 5TB, Price: 84$
	SizeSlugc4 SizeSlug = "c-4"

	// Ram: 8GB, Cpu: 4, Disk: 100GB, Transfer: 5TB, Price: 94$
	SizeSlugc24vcpu8gb SizeSlug = "c2-4vcpu-8gb"

	// Ram: 16GB, Cpu: 4, Disk: 320GB, Transfer: 8TB, Price: 96$
	SizeSlugs4vcpu16gb320gbIntel SizeSlug = "s-4vcpu-16gb-320gb-intel"

	// Ram: 16GB, Cpu: 8, Disk: 320GB, Transfer: 6TB, Price: 96$
	SizeSlugs8vcpu16gb SizeSlug = "s-8vcpu-16gb"

	// Ram: 16GB, Cpu: 2, Disk: 50GB, Transfer: 4TB, Price: 99$
	SizeSlugm2vcpu16gbIntel SizeSlug = "m-2vcpu-16gb-intel"

	// Ram: 16GB, Cpu: 2, Disk: 150GB, Transfer: 4TB, Price: 104$
	SizeSlugm32vcpu16gb SizeSlug = "m3-2vcpu-16gb"

	// Ram: 8GB, Cpu: 4, Disk: 50GB, Transfer: 5TB, Price: 109$
	SizeSlugc4Intel SizeSlug = "c-4-intel"

	// Ram: 16GB, Cpu: 2, Disk: 150GB, Transfer: 4TB, Price: 110$
	SizeSlugm32vcpu16gbIntel SizeSlug = "m3-2vcpu-16gb-intel"

	// Ram: 16GB, Cpu: 8, Disk: 320GB, Transfer: 6TB, Price: 112$
	SizeSlugs8vcpu16gbAmd SizeSlug = "s-8vcpu-16gb-amd"

	// Ram: 16GB, Cpu: 8, Disk: 320GB, Transfer: 6TB, Price: 112$
	SizeSlugs8vcpu16gbIntel SizeSlug = "s-8vcpu-16gb-intel"

	// Ram: 8GB, Cpu: 4, Disk: 100GB, Transfer: 5TB, Price: 122$
	SizeSlugc24vcpu8gbIntel SizeSlug = "c2-4vcpu-8gb-intel"

	// Ram: 16GB, Cpu: 4, Disk: 50GB, Transfer: 5TB, Price: 126$
	SizeSlugg4vcpu16gb SizeSlug = "g-4vcpu-16gb"

	// Ram: 16GB, Cpu: 8, Disk: 480GB, Transfer: 9TB, Price: 128$
	SizeSlugs8vcpu16gb480gbIntel SizeSlug = "s-8vcpu-16gb-480gb-intel"

	// Ram: 16GB, Cpu: 2, Disk: 300GB, Transfer: 4TB, Price: 131$
	SizeSlugso2vcpu16gbIntel SizeSlug = "so-2vcpu-16gb-intel"

	// Ram: 16GB, Cpu: 2, Disk: 300GB, Transfer: 4TB, Price: 131$
	SizeSlugso2vcpu16gb SizeSlug = "so-2vcpu-16gb"

	// Ram: 16GB, Cpu: 2, Disk: 300GB, Transfer: 4TB, Price: 131$
	SizeSlugm62vcpu16gb SizeSlug = "m6-2vcpu-16gb"

	// Ram: 16GB, Cpu: 4, Disk: 100GB, Transfer: 5TB, Price: 136$
	SizeSluggd4vcpu16gb SizeSlug = "gd-4vcpu-16gb"

	// Ram: 16GB, Cpu: 2, Disk: 450GB, Transfer: 4TB, Price: 139$
	SizeSlugso1_52vcpu16gbIntel SizeSlug = "so1_5-2vcpu-16gb-intel"

	// Ram: 16GB, Cpu: 4, Disk: 60GB, Transfer: 5TB, Price: 151$
	SizeSlugg4vcpu16gbIntel SizeSlug = "g-4vcpu-16gb-intel"

	// Ram: 16GB, Cpu: 4, Disk: 120GB, Transfer: 5TB, Price: 158$
	SizeSluggd4vcpu16gbIntel SizeSlug = "gd-4vcpu-16gb-intel"

	// Ram: 16GB, Cpu: 2, Disk: 450GB, Transfer: 4TB, Price: 163$
	SizeSlugso1_52vcpu16gb SizeSlug = "so1_5-2vcpu-16gb"

	// Ram: 32GB, Cpu: 8, Disk: 400GB, Transfer: 10TB, Price: 168$
	SizeSlugs8vcpu32gbAmd SizeSlug = "s-8vcpu-32gb-amd"

	// Ram: 32GB, Cpu: 4, Disk: 100GB, Transfer: 6TB, Price: 168$
	SizeSlugm4vcpu32gb SizeSlug = "m-4vcpu-32gb"

	// Ram: 16GB, Cpu: 8, Disk: 100GB, Transfer: 6TB, Price: 168$
	SizeSlugc8 SizeSlug = "c-8"

	// Ram: 16GB, Cpu: 8, Disk: 200GB, Transfer: 6TB, Price: 188$
	SizeSlugc28vcpu16gb SizeSlug = "c2-8vcpu-16gb"

	// Ram: 32GB, Cpu: 8, Disk: 640GB, Transfer: 10TB, Price: 192$
	SizeSlugs8vcpu32gb640gbIntel SizeSlug = "s-8vcpu-32gb-640gb-intel"

	// Ram: 32GB, Cpu: 4, Disk: 100GB, Transfer: 6TB, Price: 198$
	SizeSlugm4vcpu32gbIntel SizeSlug = "m-4vcpu-32gb-intel"

	// Ram: 32GB, Cpu: 4, Disk: 300GB, Transfer: 6TB, Price: 208$
	SizeSlugm34vcpu32gb SizeSlug = "m3-4vcpu-32gb"

	// Ram: 16GB, Cpu: 8, Disk: 100GB, Transfer: 6TB, Price: 218$
	SizeSlugc8Intel SizeSlug = "c-8-intel"

	// Ram: 32GB, Cpu: 4, Disk: 300GB, Transfer: 6TB, Price: 220$
	SizeSlugm34vcpu32gbIntel SizeSlug = "m3-4vcpu-32gb-intel"

	// Ram: 16GB, Cpu: 8, Disk: 200GB, Transfer: 6TB, Price: 244$
	SizeSlugc28vcpu16gbIntel SizeSlug = "c2-8vcpu-16gb-intel"

	// Ram: 32GB, Cpu: 8, Disk: 100GB, Transfer: 6TB, Price: 252$
	SizeSlugg8vcpu32gb SizeSlug = "g-8vcpu-32gb"

	// Ram: 32GB, Cpu: 4, Disk: 600GB, Transfer: 6TB, Price: 262$
	SizeSlugso4vcpu32gbIntel SizeSlug = "so-4vcpu-32gb-intel"

	// Ram: 32GB, Cpu: 4, Disk: 600GB, Transfer: 6TB, Price: 262$
	SizeSlugso4vcpu32gb SizeSlug = "so-4vcpu-32gb"

	// Ram: 32GB, Cpu: 4, Disk: 600GB, Transfer: 6TB, Price: 262$
	SizeSlugm64vcpu32gb SizeSlug = "m6-4vcpu-32gb"

	// Ram: 32GB, Cpu: 8, Disk: 200GB, Transfer: 6TB, Price: 272$
	SizeSluggd8vcpu32gb SizeSlug = "gd-8vcpu-32gb"

	// Ram: 32GB, Cpu: 4, Disk: 900GB, Transfer: 6TB, Price: 278$
	SizeSlugso1_54vcpu32gbIntel SizeSlug = "so1_5-4vcpu-32gb-intel"

	// Ram: 32GB, Cpu: 8, Disk: 120GB, Transfer: 6TB, Price: 302$
	SizeSlugg8vcpu32gbIntel SizeSlug = "g-8vcpu-32gb-intel"

	// Ram: 32GB, Cpu: 8, Disk: 240GB, Transfer: 6TB, Price: 317$
	SizeSluggd8vcpu32gbIntel SizeSlug = "gd-8vcpu-32gb-intel"

	// Ram: 32GB, Cpu: 4, Disk: 900GB, Transfer: 6TB, Price: 326$
	SizeSlugso1_54vcpu32gb SizeSlug = "so1_5-4vcpu-32gb"

	// Ram: 64GB, Cpu: 8, Disk: 200GB, Transfer: 7TB, Price: 336$
	SizeSlugm8vcpu64gb SizeSlug = "m-8vcpu-64gb"

	// Ram: 32GB, Cpu: 16, Disk: 200GB, Transfer: 7TB, Price: 336$
	SizeSlugc16 SizeSlug = "c-16"

	// Ram: 32GB, Cpu: 16, Disk: 400GB, Transfer: 7TB, Price: 376$
	SizeSlugc216vcpu32gb SizeSlug = "c2-16vcpu-32gb"

	// Ram: 64GB, Cpu: 8, Disk: 200GB, Transfer: 7TB, Price: 396$
	SizeSlugm8vcpu64gbIntel SizeSlug = "m-8vcpu-64gb-intel"

	// Ram: 64GB, Cpu: 8, Disk: 600GB, Transfer: 7TB, Price: 416$
	SizeSlugm38vcpu64gb SizeSlug = "m3-8vcpu-64gb"

	// Ram: 32GB, Cpu: 16, Disk: 200GB, Transfer: 7TB, Price: 437$
	SizeSlugc16Intel SizeSlug = "c-16-intel"

	// Ram: 64GB, Cpu: 8, Disk: 600GB, Transfer: 7TB, Price: 440$
	SizeSlugm38vcpu64gbIntel SizeSlug = "m3-8vcpu-64gb-intel"

	// Ram: 32GB, Cpu: 16, Disk: 400GB, Transfer: 7TB, Price: 489$
	SizeSlugc216vcpu32gbIntel SizeSlug = "c2-16vcpu-32gb-intel"

	// Ram: 64GB, Cpu: 16, Disk: 200GB, Transfer: 7TB, Price: 504$
	SizeSlugg16vcpu64gb SizeSlug = "g-16vcpu-64gb"

	// Ram: 64GB, Cpu: 8, Disk: 1200GB, Transfer: 7TB, Price: 524$
	SizeSlugso8vcpu64gbIntel SizeSlug = "so-8vcpu-64gb-intel"

	// Ram: 64GB, Cpu: 8, Disk: 1200GB, Transfer: 7TB, Price: 524$
	SizeSlugso8vcpu64gb SizeSlug = "so-8vcpu-64gb"

	// Ram: 64GB, Cpu: 8, Disk: 1200GB, Transfer: 7TB, Price: 524$
	SizeSlugm68vcpu64gb SizeSlug = "m6-8vcpu-64gb"

	// Ram: 64GB, Cpu: 16, Disk: 400GB, Transfer: 7TB, Price: 544$
	SizeSluggd16vcpu64gb SizeSlug = "gd-16vcpu-64gb"

	// Ram: 64GB, Cpu: 8, Disk: 1800GB, Transfer: 7TB, Price: 556$
	SizeSlugso1_58vcpu64gbIntel SizeSlug = "so1_5-8vcpu-64gb-intel"

	// Ram: 64GB, Cpu: 16, Disk: 240GB, Transfer: 7TB, Price: 605$
	SizeSlugg16vcpu64gbIntel SizeSlug = "g-16vcpu-64gb-intel"

	// Ram: 64GB, Cpu: 16, Disk: 480GB, Transfer: 7TB, Price: 634$
	SizeSluggd16vcpu64gbIntel SizeSlug = "gd-16vcpu-64gb-intel"

	// Ram: 64GB, Cpu: 8, Disk: 1800GB, Transfer: 7TB, Price: 652$
	SizeSlugso1_58vcpu64gb SizeSlug = "so1_5-8vcpu-64gb"

	// Ram: 128GB, Cpu: 16, Disk: 400GB, Transfer: 8TB, Price: 672$
	SizeSlugm16vcpu128gb SizeSlug = "m-16vcpu-128gb"

	// Ram: 64GB, Cpu: 32, Disk: 400GB, Transfer: 9TB, Price: 672$
	SizeSlugc32 SizeSlug = "c-32"

	// Ram: 64GB, Cpu: 32, Disk: 800GB, Transfer: 9TB, Price: 752$
	SizeSlugc232vcpu64gb SizeSlug = "c2-32vcpu-64gb"

	// Ram: 128GB, Cpu: 16, Disk: 400GB, Transfer: 8TB, Price: 792$
	SizeSlugm16vcpu128gbIntel SizeSlug = "m-16vcpu-128gb-intel"

	// Ram: 128GB, Cpu: 16, Disk: 1200GB, Transfer: 8TB, Price: 832$
	SizeSlugm316vcpu128gb SizeSlug = "m3-16vcpu-128gb"

	// Ram: 64GB, Cpu: 32, Disk: 400GB, Transfer: 9TB, Price: 874$
	SizeSlugc32Intel SizeSlug = "c-32-intel"

	// Ram: 128GB, Cpu: 16, Disk: 1200GB, Transfer: 8TB, Price: 880$
	SizeSlugm316vcpu128gbIntel SizeSlug = "m3-16vcpu-128gb-intel"

	// Ram: 64GB, Cpu: 32, Disk: 800GB, Transfer: 9TB, Price: 978$
	SizeSlugc232vcpu64gbIntel SizeSlug = "c2-32vcpu-64gb-intel"

	// Ram: 96GB, Cpu: 48, Disk: 600GB, Transfer: 11TB, Price: 1008$
	SizeSlugc48 SizeSlug = "c-48"

	// Ram: 192GB, Cpu: 24, Disk: 600GB, Transfer: 9TB, Price: 1008$
	SizeSlugm24vcpu192gb SizeSlug = "m-24vcpu-192gb"

	// Ram: 128GB, Cpu: 32, Disk: 400GB, Transfer: 8TB, Price: 1008$
	SizeSlugg32vcpu128gb SizeSlug = "g-32vcpu-128gb"

	// Ram: 128GB, Cpu: 16, Disk: 2400GB, Transfer: 8TB, Price: 1048$
	SizeSlugso16vcpu128gbIntel SizeSlug = "so-16vcpu-128gb-intel"

	// Ram: 128GB, Cpu: 16, Disk: 2400GB, Transfer: 8TB, Price: 1048$
	SizeSlugso16vcpu128gb SizeSlug = "so-16vcpu-128gb"

	// Ram: 128GB, Cpu: 16, Disk: 2400GB, Transfer: 8TB, Price: 1048$
	SizeSlugm616vcpu128gb SizeSlug = "m6-16vcpu-128gb"

	// Ram: 128GB, Cpu: 32, Disk: 800GB, Transfer: 8TB, Price: 1088$
	SizeSluggd32vcpu128gb SizeSlug = "gd-32vcpu-128gb"

	// Ram: 128GB, Cpu: 16, Disk: 3600GB, Transfer: 8TB, Price: 1112$
	SizeSlugso1_516vcpu128gbIntel SizeSlug = "so1_5-16vcpu-128gb-intel"

	// Ram: 96GB, Cpu: 48, Disk: 1200GB, Transfer: 11TB, Price: 1128$
	SizeSlugc248vcpu96gb SizeSlug = "c2-48vcpu-96gb"

	// Ram: 192GB, Cpu: 24, Disk: 600GB, Transfer: 9TB, Price: 1188$
	SizeSlugm24vcpu192gbIntel SizeSlug = "m-24vcpu-192gb-intel"

	// Ram: 128GB, Cpu: 32, Disk: 480GB, Transfer: 8TB, Price: 1210$
	SizeSlugg32vcpu128gbIntel SizeSlug = "g-32vcpu-128gb-intel"

	// Ram: 192GB, Cpu: 24, Disk: 1800GB, Transfer: 9TB, Price: 1248$
	SizeSlugm324vcpu192gb SizeSlug = "m3-24vcpu-192gb"

	// Ram: 160GB, Cpu: 40, Disk: 500GB, Transfer: 9TB, Price: 1260$
	SizeSlugg40vcpu160gb SizeSlug = "g-40vcpu-160gb"

	// Ram: 128GB, Cpu: 32, Disk: 960GB, Transfer: 8TB, Price: 1268$
	SizeSluggd32vcpu128gbIntel SizeSlug = "gd-32vcpu-128gb-intel"

	// Ram: 128GB, Cpu: 16, Disk: 3600GB, Transfer: 8TB, Price: 1304$
	SizeSlugso1_516vcpu128gb SizeSlug = "so1_5-16vcpu-128gb"

	// Ram: 96GB, Cpu: 48, Disk: 600GB, Transfer: 11TB, Price: 1310$
	SizeSlugc48Intel SizeSlug = "c-48-intel"

	// Ram: 192GB, Cpu: 24, Disk: 1800GB, Transfer: 9TB, Price: 1320$
	SizeSlugm324vcpu192gbIntel SizeSlug = "m3-24vcpu-192gb-intel"

	// Ram: 256GB, Cpu: 32, Disk: 800GB, Transfer: 10TB, Price: 1344$
	SizeSlugm32vcpu256gb SizeSlug = "m-32vcpu-256gb"

	// Ram: 160GB, Cpu: 40, Disk: 1000GB, Transfer: 9TB, Price: 1360$
	SizeSluggd40vcpu160gb SizeSlug = "gd-40vcpu-160gb"

	// Ram: 96GB, Cpu: 48, Disk: 1200GB, Transfer: 11TB, Price: 1466$
	SizeSlugc248vcpu96gbIntel SizeSlug = "c2-48vcpu-96gb-intel"

	// Ram: 192GB, Cpu: 24, Disk: 3600GB, Transfer: 9TB, Price: 1572$
	SizeSlugso24vcpu192gbIntel SizeSlug = "so-24vcpu-192gb-intel"

	// Ram: 192GB, Cpu: 24, Disk: 3600GB, Transfer: 9TB, Price: 1572$
	SizeSlugso24vcpu192gb SizeSlug = "so-24vcpu-192gb"

	// Ram: 192GB, Cpu: 24, Disk: 3600GB, Transfer: 9TB, Price: 1572$
	SizeSlugm624vcpu192gb SizeSlug = "m6-24vcpu-192gb"

	// Ram: 256GB, Cpu: 32, Disk: 800GB, Transfer: 10TB, Price: 1584$
	SizeSlugm32vcpu256gbIntel SizeSlug = "m-32vcpu-256gb-intel"

	// Ram: 256GB, Cpu: 32, Disk: 2400GB, Transfer: 10TB, Price: 1664$
	SizeSlugm332vcpu256gb SizeSlug = "m3-32vcpu-256gb"

	// Ram: 192GB, Cpu: 24, Disk: 5400GB, Transfer: 9TB, Price: 1668$
	SizeSlugso1_524vcpu192gbIntel SizeSlug = "so1_5-24vcpu-192gb-intel"

	// Ram: 256GB, Cpu: 32, Disk: 2400GB, Transfer: 10TB, Price: 1760$
	SizeSlugm332vcpu256gbIntel SizeSlug = "m3-32vcpu-256gb-intel"

	// Ram: 192GB, Cpu: 48, Disk: 720GB, Transfer: 9TB, Price: 1814$
	SizeSlugg48vcpu192gbIntel SizeSlug = "g-48vcpu-192gb-intel"

	// Ram: 192GB, Cpu: 48, Disk: 1440GB, Transfer: 11TB, Price: 1901$
	SizeSluggd48vcpu192gbIntel SizeSlug = "gd-48vcpu-192gb-intel"

	// Ram: 192GB, Cpu: 24, Disk: 5400GB, Transfer: 9TB, Price: 1956$
	SizeSlugso1_524vcpu192gb SizeSlug = "so1_5-24vcpu-192gb"

	// Ram: 256GB, Cpu: 32, Disk: 4800GB, Transfer: 10TB, Price: 2096$
	SizeSlugso32vcpu256gbIntel SizeSlug = "so-32vcpu-256gb-intel"

	// Ram: 256GB, Cpu: 32, Disk: 4800GB, Transfer: 10TB, Price: 2096$
	SizeSlugso32vcpu256gb SizeSlug = "so-32vcpu-256gb"

	// Ram: 256GB, Cpu: 32, Disk: 4800GB, Transfer: 10TB, Price: 2096$
	SizeSlugm632vcpu256gb SizeSlug = "m6-32vcpu-256gb"

	// Ram: 256GB, Cpu: 32, Disk: 7200GB, Transfer: 10TB, Price: 2224$
	SizeSlugso1_532vcpu256gbIntel SizeSlug = "so1_5-32vcpu-256gb-intel"

	// Ram: 256GB, Cpu: 32, Disk: 7200GB, Transfer: 10TB, Price: 2608$
	SizeSlugso1_532vcpu256gb SizeSlug = "so1_5-32vcpu-256gb"
)
