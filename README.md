# bachelor thesis image matching



# Installation
- image_matcher/image_matcher is a pre compiled binary which can be executed without installing anything

- if you want to build the project you need to install go https://go.dev/doc/install
- `cd image_matcher` and execute `go get` to install dependencies from the go.mod
- run `go build` to build the project 

*Starting database*

- you can start the mysql database by simply executing `docker-compose up` command in projects root

# Arguments
*`image_path`:*
- relative path to an image
- should always include .png at the end

*`<analyzer>`: sift | orb | brisk | phash | new*

- sift, orb, brisk always need the `<matcher>` argument when running commands
- phash and new don't need the `<matcher>` when running commands
- new is the new algorithm implemented for the bachelors thesis
- the local phash implementation doesn't work as well as the phash-calculator from spreadshirt
  - for many images there is a 0 hash evaluated 
  - this has a big impact on matching performance for both phash and the new algorithm

*`<matcher>`: bfm | flann*

*`<threshold>`:*
- values between 0 and 1 for sift, orb and brisk
- integer values >= 0 for phash and new

*`<scenario>`: identical | scaled | rotated | background | mirrored | moved | part | mixed | all*

# Commands
*`./image_matcher compare <image1_path> <image2_path> <analyzer> <matcher> <threshold>`*
- matches two specified images with each other
- threshold argument is optional
- database is not needed

*`./image_matcher register <directory_path | image_path>`*
- registers an image in the forbidden set in the database
- argument can be path to a directory, to save multiple images at once
- the images are not saved in the db, only the descriptors and hash values are stored

*`image_matcher/image_matcher duplicate <directory_path>`*
- generates modified duplicates from the originals and stores them in the database as search images
- `<directory_path>` should be path to the images that were registered with the `register` command
- **the images are not saved in the db. they are generated and saved in `images/variations/`**
- **the search images are expected to be found in images/variations when running a scenario**
- **command should be run from project root**

*`image_matcher/image_matcher uniques <directory_path>`*
- generates unique images and stores them in the database as search images
- `<directory_path>` **should not** be a path to images that were already registered in the forbidden set in the 
  duplicate set
- **the images are not saved in the db. they are generated and saved in `images/variations/`**
- **the search images are expected to be found in images/variations when running a scenario**
- **command should be run from project root**

*`./image_matcher match <image_path> <analyzer> <matcher> <thresholdK>`*
- matches image from path against database
- threshold argument is optional

*`image_matcher/image_matcher scenario <scenario> <analyzer> <matcher> <threshold>`*
- runs the specified scenario for the algorithm
- the results from the tests are saved in test-output/csv-files
- **the search images are expected to be found in images/variations when running a scenario**
- **command should be run from project root**

*`image_matcher/image_matcher runAll`*
- runs all scenarios for phash, sift, brisk and orb
- the results from the tests are saved in test-output/csv-files
- **the search images are expected to be found in images/variations when running a scenario**
- **command should be run from project root**
