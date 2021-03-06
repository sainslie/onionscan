# What is scanned for?

Listed below are a few of the more serious privacy problems that may be found
during a scan, ordered per scan type.

## Web sites

When OnionScan detects a web server, it is scanned for the issues described in this section.

### Apache mod_status Protection

This [should not be news](http://arstechnica.com/security/2016/02/default-settings-in-apache-may-decloak-tor-hidden-services/), you should not have it enabled. If you do have it enabled, attacks can:

* Build a better fingerprint of your server, including php and other software versions.
* Determine client IP addresses if you are co-hosting a clearnet site.
* Determine your IP address if your setup allows.
* Determine other sites you are co-hosting.
* Determine how active your site is.
* Find secret or hidden areas of your site
* and much, much more.

Seriously, don't even run the tool, go to your site and check if you have `/server-status`
reachable. If you do, turn it off!

### Open Directories

Basic web security 101, if you leave directories open then people are going to scan
them, and find interesting things - old versions of images, temp files etc.

Many sites use common structures `style/`, `images/` etc. The tool checks for
common variations, and allows the user to submit others for testing. 

### EXIF Tags

Whether you create them yourself or allow users to upload images, you need to
ensure the metadata associated with the image is stripped.

Many, many websites still do not properly sanitise image data, leaving themselves
or their users at risk of deanonymization.

### Server Fingerprint

Sometimes, even without mod_status we can determine if two sites are hosted on
 the same infrastructure. We can use the following attributes to make this distinction:

* Server HTTP Header
* Technology Stack (e.g. php, jquery version etc.)
* Website folder layout e.g. do you use `/style` or `/css` or do you use wordpress.
* Fingerprints of images
* GPG Versions being used.
