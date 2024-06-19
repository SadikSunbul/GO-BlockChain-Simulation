package blockchain

type TxOutput struct { //transectıon cıktıları
	Value  int    //token degeri
	PubKey string //publıkkey sonra burası degısıcektır suan pubkey yerıne herhangıbır strıng deger kullanılıcak
}

type TxInput struct { //transectıon girdileri
	ID  []byte //cıkısı referans eder
	Out int    //cıkıs endexı  referans eder
	Sig string //gırıs verısıdir
}

/*
CanUnlock metodunun görevi, bir işlem girişinin belirli bir veri ile kilidini açıp açamayacağını kontrol etmektir.
Genellikle işlem girişleri, işlemi imzalayan kişinin imzasını içerir. Bu metod, girişin imza alanının belirli bir
veri ile eşleşip eşleşmediğini kontrol eder. Eğer eşleşiyorsa, girişin doğru kişi tarafından yapıldığı doğrulanmış olur.
*/
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
	// Bu fonksiyon, bir işlem girişinin belirli bir veri ile kilidini açıp açamayacağını kontrol eder.
	// Girişin imza (Sig) alanı, verilen data değeri ile eşleşiyorsa true döner.
	// Bu, girişin sahibinin işlemi imzalayan doğru kişi olduğunu doğrular.
}

/*
CanBeUnlocked metodunun amacı, bir işlem çıkışının belirli bir veri ile kilidini açıp açamayacağını kontrol etmektir.
Çıkış genellikle genel anahtar (public key) ile kilitlenmiştir ve bu metod, çıkışın belirli bir genel anahtar ile
eşleşip eşleşmediğini kontrol eder. Eğer eşleşiyorsa, çıkışın doğru kişiye ait olduğu doğrulanmış olur.
*/
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
	// Bu fonksiyon, bir işlem çıkışının belirli bir veri ile kilidini açıp açamayacağını kontrol eder.
	// Çıkışın genel anahtarı (PubKey), verilen data değeri ile eşleşiyorsa true döner.
	// Bu, çıkışın belirli bir kişiye ait olduğunu doğrular.
}
