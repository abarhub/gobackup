package hashFiles

import (
	"crypto/sha256"
	"gobackup/internal/config"
	"gobackup/internal/utils"
	"hash"
	"os"
	"testing"
)

func TestConstruitHash(t *testing.T) {
	type args struct {
		contenuFichier string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		hash    string
	}{
		{name: "test1", args: args{contenuFichier: "test1"}, wantErr: false,
			hash: "1b4f0e9851971998e732078544c96b36c3d01cedf7caa332359d6f1d83567014"},
		{name: "test2", args: args{contenuFichier: "abc123   ZZZZZWWWW_%xyztt"}, wantErr: false,
			hash: "862c8c56297f6e18a692c548c4918e71a7b626e55e9f272c6248670fc9388779"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := t.TempDir() + "/GeeksforGeeks.txt"
			err := utils.CreateFile(f, tt.args.contenuFichier)
			if err != nil {
				t.Errorf("error for create file %s : %v", f, err)
			} else {
				if err := ConstruitHash(f); (err != nil) != tt.wantErr {
					t.Errorf("ConstruitHash() error = %v, wantErr %v", err, tt.wantErr)
				} else {
					s, err := utils.ReadFile(f + GetExtension())
					if err != nil {
						t.Errorf("erreur pour lire le fichier %s : %v", f, err)
					} else if s != tt.hash {
						t.Errorf("les hash sont différents %s != %s : %v", s, tt.hash, err)
					}
				}
			}
		})
	}
}

func TestGetExtension(t *testing.T) {
	want := ".sha256sum"
	if got := GetExtension(); got != want {
		t.Errorf("GetExtension() = %v, want %v", got, want)
	}
}

func TestVerifieHash(t *testing.T) {
	type args struct {
		contenuFichier string
		config         config.BackupGlobal
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test1", args: args{contenuFichier: "abc123"}, wantErr: false},
		{name: "test2", args: args{contenuFichier: "xxxYYY_"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep0 := t.TempDir()
			nom := "rep1"
			rep := rep0 + "/" + nom
			err := os.MkdirAll(rep, os.ModePerm)
			if err != nil {
				t.Errorf("error for create file %s : %v", rep, err)
			} else {
				configBackup := config.BackupGlobal{}
				configBackup.RepCompression = rep0
				configBackup.RepCryptage = rep0
				f := rep + "/test1.txt"
				err := utils.CreateFile(f, tt.args.contenuFichier)
				if err != nil {
					t.Errorf("error for create file %s : %v", f, err)
				} else if err := ConstruitHash(f); (err != nil) != tt.wantErr {
					t.Errorf("ConstruitHash() error = %v, wantErr %v", err, tt.wantErr)
				} else if err := VerifieHash(nom, configBackup); (err != nil) != tt.wantErr {
					t.Errorf("VerifieHash() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func Test_calculHash(t *testing.T) {
	type args struct {
		contenuFichier string
		hasher         hash.Hash
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "test1", args: args{contenuFichier: "abc123"},
			want: "6ca13d52ca70c883e0f0bb101e425a89e8624de51db2d2392593af6a84118090", wantErr: false},
		{name: "test2", args: args{contenuFichier: "xxxyyyZZZ_123"},
			want: "6b164129d4ec61d9fbaef5d1cd66738070b49dd7c0694b7e3aab13bfcbf7e0ea", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := t.TempDir() + "/test1.txt"
			err := utils.CreateFile(f, tt.args.contenuFichier)
			if err != nil {
				t.Errorf("error for create file %s : %v", f, err)
			} else {
				var hashAlgo = sha256.New()
				got, err := calculHash(f, hashAlgo)
				if (err != nil) != tt.wantErr {
					t.Errorf("calculHash() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("calculHash() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_verifieFichiers(t *testing.T) {
	type args struct {
		nomFichier     string
		contenuFichier string
		contenuHash    string
		sansHash       bool
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		nbIgnore int
	}{
		{name: "test1", args: args{contenuFichier: "xxx",
			contenuHash: "cd2eb0837c9b4c962c22d2ff8b5441b7b45805887f051d39bf133b583baf6860", sansHash: false}, wantErr: false},
		{name: "test2", args: args{contenuFichier: "abc123_ZZZ",
			contenuHash: "7faaf88ac147ce0384df305c4263c6dae094797b90586048749b691b0181d554", sansHash: false}, wantErr: false},
		{name: "test3", args: args{contenuFichier: "",
			contenuHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", sansHash: false}, wantErr: false},
		{name: "test4", args: args{contenuFichier: "ZZZ", sansHash: true}, wantErr: false, nbIgnore: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := t.TempDir()
			fichier := rep + "/test1.txt"
			err := utils.CreateFile(fichier, tt.args.contenuFichier)
			if err != nil {
				t.Errorf("error for create file %s : %v", fichier, err)
			} else {
				if !tt.args.sansHash {
					err = utils.CreateFile(fichier+GetExtension(), tt.args.contenuHash)
				}
				if err != nil {
					t.Errorf("error for create file hash %s : %v", fichier+GetExtension(), err)
				} else {
					if nbIgnore, err := verifieFichiers(rep); (err != nil) != tt.wantErr {
						t.Errorf("verifieFichiers() error = %v, wantErr %v", err, tt.wantErr)
					} else if nbIgnore != tt.nbIgnore {
						t.Errorf("verifieFichiers() nb ignore = %d", nbIgnore)
					}
				}
			}
		})
	}
}
