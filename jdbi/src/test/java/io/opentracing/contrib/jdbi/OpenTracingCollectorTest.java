package io.opentracing.contrib.jdbi;

import com.lightstep.tracer.jre.JRETracer;
import com.lightstep.tracer.shared.Options;
import io.opentracing.Span;
import io.opentracing.Tracer;
import org.junit.Test;
import org.skife.jdbi.v2.*;

import java.util.List;
import java.util.Map;

public class OpenTracingCollectorTest {
    @Test
    public void testBasics() throws ClassNotFoundException {
        Class.forName("com.mysql.jdbc.Driver");
        Tracer tracer = new JRETracer(
                new Options("XXXX")
                        .withCollectorHost("collector-staging.lightstep.com"));
        try (Span parent = tracer.buildSpan("parent tester").start()) {
            DBI dbi = new DBI("jdbc:mysql://localhost/crouton", "root", "");
            dbi.setTimingCollector(new OpenTracingCollector(tracer));
            System.out.println(dbi);
            Handle handle = dbi.open();

            Query<Map<String, Object>> statement = handle.createQuery("SELECT COUNT(*) FROM projects");
            OpenTracingCollector.setParent(statement, parent);

            List<Map<String, Object>> results = statement.list();
            System.out.println("results" + results);
        }
    }
}
