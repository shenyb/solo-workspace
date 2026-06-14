package {{basePackage}}.common.api;

import lombok.Getter;
import java.util.List;

@Getter
public class PageResult<T> {

    private final long total;
    private final int page;
    private final int size;
    private final List<T> records;

    public PageResult(long total, int page, int size, List<T> records) {
        this.total = total;
        this.page = page;
        this.size = size;
        this.records = records;
    }

    public static <T> PageResult<T> of(long total, int page, int size, List<T> records) {
        return new PageResult<>(total, page, size, records);
    }
}
