# hostfilter
A small program to filter ads, trackers and malware using your systems
hosts-file.

# About the filtering
There are multiple ways to filter ads, trackers or malwaresites. There
are bowserplugins, firewalls etc. than can help you doing the job. One
nice way is to filter the bad domainnames using your systems hosts-file.
This way all internet-traffic of your machine is protected against
those known bad guys.
While this approach can only filter traffic to known domainnames it is
no 100% protection. But with ease you can block traffic to over 100000
known bad sites.
This way you don't need any additional software, because the hosts-file
is part of your OS.

# How hostfilter works
hostfiler downloads known hosts-files, containing ad, tracker and
malwaresites and put them inside your systems hosts-file. This way you
can prevent all communication with those sites.
Running hostfilter downloads the ad-hosts-files you specified and adds
them to your systems hosts-file. As it is unlikely that a malwaredomain
will become a nice place to browse, domains that were added to your
hosts-file, even if they are no longer inside the downloaded
hosts-files. Domainnames are only added once, so you can run hostfilter
multiple times without bloating you systems hosts-file.
Running hostfilter the first time should create a backup of your current
systems hosts-file.
hostfilter does not interact in any way with your networktraffic. It
just puts malwaredomains in your hosts-file and tells your system to
discard all traffic to them.

# Configuration
Only configuration you can do is to set the URLs to the ad-hosts-files
containing the domains to block. hostfilter searches in the current
directory and in your etc-directory for a file called adhosts.cfg which
contains the URLs to the ad-hosts-files to use. The repository contains an
example-file you can use, that holds a set of URLs to filter more
than 116000 domains.

# hosts-files
There are two types of ad-hosts-files that are supported. The first kind is
structured like a normal hosts-file, so it has an IP and a domainname
per line. To block the domains the IP should point to 127.0.0.1 or to
0.0.0.0. This kind of file looks the following way.

    127.0.0.1 malwaredomain.com
    127.0.0.1 anotherdomain.com

There are only domains added to the blacklist that have the IP 127.0.0.1
or 0.0.0.0 attached, any other host will be ignored. This is done to
prevent that a bad hosts-file redirects your traffic.

The second type of supported files contains only a domainname per line.
Those files look like this.

    mailwaredomain.com
    anotherdomain.com

