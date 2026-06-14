package {{basePackage}}.web;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication(scanBasePackages = "{{basePackage}}")
@MapperScan(basePackages = "{{basePackage}}", annotationClass = org.apache.ibatis.annotations.Mapper.class)
public class {{ProjectName}}Application {

    public static void main(String[] args) {
        SpringApplication.run({{ProjectName}}Application.class, args);
    }
}
