package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gnames/gnames/config"
)

var _ = Describe("Config", func() {
	Describe("NewConfig", func() {
		It("Creates a default GNparser", func() {
			cnf := NewConfig()
			Expect(cnf.JobsNum).To(Equal(8))
			Expect(cnf.PgHost).To(Equal("localhost"))
			deflt := Config{
				WorkDir:     "/tmp/gnmatcher",
				JobsNum:     8,
				MaxEditDist: 1,
				PgHost:      "localhost",
				PgPort:      5432,
				PgUser:      "postgres",
				PgPass:      "",
				PgDB:        "gnames",
			}
			Expect(cnf).To(Equal(deflt))
		})
	})

	It("Takes options to update default settings", func() {
		opts := opts()
		cnf := NewConfig(opts...)
		updt := Config{
			WorkDir:     "/var/opt/gnmatcher",
			JobsNum:     16,
			MaxEditDist: 2,
			PgHost:      "mypg",
			PgPort:      1234,
			PgUser:      "gnm",
			PgPass:      "secret",
			PgDB:        "gnm",
		}
		Expect(cnf).To(Equal(updt))
	})

	It("It limits MaxEditDist to 1 and 2", func() {
		cnf := NewConfig(OptMaxEditDist(5))
		Expect(cnf.MaxEditDist).To(Equal(1))
		cnf = NewConfig(OptMaxEditDist(0))
		Expect(cnf.MaxEditDist).To(Equal(1))
	})
})

func opts() []Option {
	return []Option{
		OptWorkDir("/var/opt/gnmatcher"),
		OptJobsNum(16),
		OptMaxEditDist(2),
		OptPgHost("mypg"),
		OptPgUser("gnm"),
		OptPgPass("secret"),
		OptPgPort(1234),
		OptPgDB("gnm"),
	}
}
