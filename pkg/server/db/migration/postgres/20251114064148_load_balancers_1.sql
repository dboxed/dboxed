-- +goose Up
ALTER TABLE ingress_proxy RENAME TO load_balancer;
ALTER TABLE ingress_proxy_box RENAME TO load_balancer_box;
ALTER TABLE box_ingress RENAME TO load_balancer_service;
