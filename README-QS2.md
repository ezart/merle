## Quick Start, Part II

In Part I, we ran Merle locally on a Raspberry Pi and used a web browser to view the Thing on http://localhost.  In Part II, we're going to run Merle on another system on the Internet.  This second system runs a copy of the Thing.  We'll call this copy Thing' (Thing prime).  Thing and Thing' are connected to each other over a secure tunnel.  

<img src="https://docs.google.com/drawings/d/e/2PACX-1vTnTWN_AcEEWgGjuN7uF13Z3lCikFMPzGR7eSovbkdgLS0__YcJ5Azh6BWQZLchuh12HZahR8VyH05F/pub?w=2023&amp;h=921">

Thing and Thing' are syncronized.  Any state change in Thing is reflected in Thing', and visa-versa.  Thing' view through a web browser shows the same web-app view we see when viewing Thing.  In fact, the same front-end code (HTML, CSS, Javascript) and back-end code (Go) is running in both places.  But notice Thing' does not have access to hardware.  Thing' is a proxy for Thing's hardware.

