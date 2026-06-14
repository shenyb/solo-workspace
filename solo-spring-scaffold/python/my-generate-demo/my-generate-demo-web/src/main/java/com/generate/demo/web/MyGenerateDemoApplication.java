package com.generate.demo.web;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication(scanBasePackages = "com.generate.demo")
@MapperScan(basePackages = "com.generate.demo", annotationClass = org.apache.ibatis.annotations.Mapper.class)
public class MyGenerateDemoApplication {

    public static void main(String[] args) {
        SpringApplication.run(MyGenerateDemoApplication.class, args);
    }
}
