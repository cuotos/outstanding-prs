FROM alpine
COPY outstanding-prs /
ENTRYPOINT ["/outstanding-prs"]