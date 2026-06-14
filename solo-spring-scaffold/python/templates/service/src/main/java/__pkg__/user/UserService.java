package {{basePackage}}.user;

import {{basePackage}}.common.api.PageResult;
import {{basePackage}}.user.dto.UserCreateRequest;
import {{basePackage}}.user.dto.UserResponse;
import {{basePackage}}.user.dto.UserUpdateRequest;

public interface UserService {

    UserResponse getById(Long id);

    PageResult<UserResponse> page(int page, int size);

    Long create(UserCreateRequest request);

    void update(UserUpdateRequest request);

    void delete(Long id);
}
