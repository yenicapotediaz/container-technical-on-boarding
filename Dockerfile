FROM golang:1.9

ARG VERSION
ARG BUILD
ENV PACKAGE_PATH "/go/src/github.com/samsung-cnct/container-technical-on-boarding"
ENV ONBOARD_TASKS_FILE "/workload/onboarding-issues.yaml"

RUN apt-get -qq update && apt-get install -y -q build-essential

WORKDIR ${PACKAGE_PATH}
COPY . ${PACKAGE_PATH}

# set version
RUN sed -i -- 's/${VERSION}/'"$VERSION"'/g' conf/app.conf && \
	  sed -i -- 's/${BUILD}/'"$BUILD"'/g' conf/app.conf

RUN make all
RUN mkdir /workload && \
    cp -v ${PACKAGE_PATH}/onboarding/onboarding-issues.yaml ${ONBOARD_TASKS_FILE}

VOLUME ["/go/"]
EXPOSE 9000

CMD ["revel", "run", "github.com/samsung-cnct/container-technical-on-boarding"]
