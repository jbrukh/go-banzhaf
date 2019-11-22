# go-banzhaf

Go implementation of Banzhaf power index calculation.

# Background

The Banzhaf power index is one way to measure voting power in a weighted voting system. This package provides an algorithm which calculates absolute and normalized Banzhaf voting power indices.

# Usage

Given a weighted voting system with a quota and weights, use the `Banzhaf` function to get a list of power index calculations.

    weights := []uint64{2, 2, 2, 1}
    quota := uint64(4)
    absolute := true
    
    index, err := Banzhaf(weights, quota, absolute)
    if err != nil {
      // error
    }

# References

* [Are blockchain voters 'dummies'?](https://blog.coinfund.io/are-blockchain-voters-dummies-4a89a376de69) by @jbrukh.
* [Using generator functions to compute power indices](http://www.siue.edu/~aweyhau/teaching/seniorprojects/heger_final.pdf) by Brian Hegers
* https://gist.github.com/HeinrichHartmann/8ec2e2245f2a70441257 by Heinrich Hartmann

