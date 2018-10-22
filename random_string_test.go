package main

import "testing"

func TestSimplifySuffixSamePrefix(t *testing.T) {
	if v := simplifySuffix("aaddaa"); v != "aadda" {
		t.Log(v)
		t.Fail()
	}
}

func TestSimplifySuffixAllTheSame(t *testing.T) {
	if v := simplifySuffix("aaaa"); v != "a" {
		t.Log(v)
		t.Fail()
	}
}

func TestSimplifySuffixEmpty(t *testing.T) {
	if v := simplifySuffix(""); v != "" {
		t.Log(v)
		t.Fail()
	}
}

func TestSimplifySuffixOneLetter(t *testing.T) {
	if v := simplifySuffix("a"); v != "a" {
		t.Log(v)
		t.Fail()
	}
}

func TestSimplifySuffixBigSuffix(t *testing.T) {
	if v := simplifySuffix("hfaaaaa"); v != "hfa" {
		t.Log(v)
		t.Fail()
	}
}
