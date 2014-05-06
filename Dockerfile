FROM fabiokung/cedar

RUN curl http://naaman-scratch.s3.amazonaws.com/bp -O -L
RUN chmod +x bp
RUN mv bp /usr/local/bin

RUN mkdir -p /opt/buildpacks
RUN git clone git://github.com/heroku/heroku-buildpack-ruby.git /opt/buildpacks/ruby

