#!/bin/bash

function main() {
  bash files/1_web_http.sh
  bash files/2_micro_grpc.sh
  bash files/3_micro_grpc_http.sh
  bash files/4_web_http_pb.sh
  bash files/5_micro_grpc_pb.sh
  bash files/6_micro_grpc_http_pb.sh
  bash files/7_micro_grpc_gateway_pb.sh
}

main
